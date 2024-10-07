package s3

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
)

type PhotoFile struct {
	Name string

	UserUUID uuid.UUID

	LocalPath string
	SaveTo    string
	Ext       string
	Width     int
	Height    int
	Resized   bool

	ContentType string
	Size        int64
}

func NewPhotoFile(path string, userUUID uuid.UUID, width int) (pf *PhotoFile, err error) {
	mime, err := helpers.FileMimetype(path)
	if err != nil {
		return pf, err
	}

	args := ""
	if width != OriginalPhotoSize {
		args = fmt.Sprintf("w%d", width) // w100
	}

	size, err := helpers.FileSize(path)
	if err != nil {
		return pf, err
	}

	ext := helpers.FileExt(path)
	rwidth, rheight, err := helpers.ImageSize(path)
	if err != nil {
		return pf, err
	}

	// @todo: convert to jpg
	objectPath := fmt.Sprintf("photos-%s.%s.jpg", userUUID, args)
	objectPath = strings.ReplaceAll(objectPath, "..", ".")
	objectPath = strings.ReplaceAll(objectPath, "//", "/")

	return &PhotoFile{
		UserUUID:    userUUID,
		Name:        helpers.ParsePathFileName(path),
		LocalPath:   path,
		SaveTo:      objectPath,
		Ext:         ext,
		Width:       rwidth,
		Height:      rheight,
		ContentType: mime,
		Size:        size,
		Resized:     width != OriginalPhotoSize,
	}, nil
}

type FileDTO struct {
	Name       string
	ObjectName string

	UserUUID uuid.UUID

	LocalPath string
	SaveTo    string
	Ext       string

	Width  int
	Height int

	ContentType string
	Size        int64
}

func NewFileDTO(fileName, path, objectName string, userUUID uuid.UUID) (pf *FileDTO, err error) {
	mime, err := helpers.FileMimetype(path)
	if err != nil {
		return pf, err
	}

	name := helpers.ParsePathFileName(path)
	ext := helpers.FileExt(path)

	size, err := helpers.FileSize(path)
	if err != nil {
		return pf, err
	}

	width, height := 0, 0
	if helpers.FileMimeIsImage(mime) {
		width, height, err = helpers.ImageSize(path)
		if err != nil {
			return pf, err
		}
	}

	objectPath := fmt.Sprintf("%s.%s", name, ext)
	objectPath = strings.ReplaceAll(objectPath, "..", ".")
	objectPath = strings.ReplaceAll(objectPath, "//", "/")

	return &FileDTO{
		UserUUID:   userUUID,
		Name:       fileName,
		ObjectName: objectName,
		LocalPath:  path,
		SaveTo:     objectPath,
		Ext:        ext,

		Width:  width,
		Height: height,

		ContentType: mime,
		Size:        size,
	}, nil
}
