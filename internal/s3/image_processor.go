package s3

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/sirupsen/logrus"
)

type ToUpload struct {
	Ctx          context.Context
	Path         string
	OriginalPath string
	UserUUID     uuid.UUID
	Width        int

	Done *sync.WaitGroup
}

func (s3 *Service) ToProcess(up ToUpload) {
	s3.toResize <- up
}

func (s3 *Service) ToResize() {
	for v := range s3.toResize {
		ctx, span := NewSpanWithCtx(v.Ctx, fmt.Sprintf("Resize%d", v.Width))
		v.Ctx = ctx

		if v.Width != OriginalPhotoSize {
			logrus.Infof("resizing... %v %v: ", v.Path, v.Width)

			resizedPath, err := helpers.ResizeImage(v.Path, v.Width)
			if err != nil {
				s3.toError <- err.Error()
				continue
			}
			logrus.Info("resized: ", resizedPath)
			v.Path = resizedPath
		}
		Span(span)()

		s3.toUpload <- v
	}
}

func (s3 *Service) ToUpload() {
	for v := range s3.toUpload {

		span := NewSpan(v.Ctx, fmt.Sprintf("Upload%d", v.Width))

		logrus.Infof("uploading... %s %v", v.Path, v.Width)
		photoFile, err := NewPhotoFile(v.Path, v.UserUUID, v.Width)
		if err != nil {
			s3.toError <- err.Error()
			v.Done.Done()
			continue
		}

		err = s3.Upload(photoFile.LocalPath, photoFile.ContentType, photoFile.SaveTo)
		if err != nil {
			s3.toError <- err.Error()
			v.Done.Done()
			continue
		}

		logrus.Infof("uploaded: %s %v", v.Path, v.Width)

		s3.toUploadedDone <- *photoFile
		v.Done.Done()

		Span(span)()
	}
}
