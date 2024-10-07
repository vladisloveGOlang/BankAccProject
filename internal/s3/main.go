package s3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	location        string // "ru-central1"
	useSSL          bool
	publicURL       string

	repo *Repository

	cropWidth []int

	toResize       chan ToUpload
	toUpload       chan ToUpload
	toError        chan string
	toUploadedDone chan PhotoFile

	UploadTimeout  time.Duration
	ParallelUpload int
	ParallelResize int
}

const (
	OriginalPhotoSize = 0
	SmallPhotoSize    = 50
	MediumPhotoSize   = 200
	LargePhotoSize    = 600
)

type Conf struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Location        string
	UseSSL          bool
	PublicURL       string
}

func New(conf Conf, repo *Repository) *Service {
	s3 := &Service{
		endpoint:        conf.Endpoint,
		accessKeyID:     conf.AccessKeyID,
		secretAccessKey: conf.SecretAccessKey,
		bucketName:      conf.BucketName,
		location:        conf.Location,
		useSSL:          conf.UseSSL,
		publicURL:       conf.PublicURL,
		repo:            repo,
		cropWidth:       []int{OriginalPhotoSize, SmallPhotoSize, LargePhotoSize, MediumPhotoSize},

		toResize:       make(chan ToUpload, 1000),
		toUpload:       make(chan ToUpload, 1000),
		toUploadedDone: make(chan PhotoFile, 1000),
		toError:        make(chan string, 1000),

		UploadTimeout:  90 * time.Second,
		ParallelUpload: 12,
		ParallelResize: 4,
	}

	go func() {
		for e := range s3.toError {
			logrus.Error(e)
		}
	}()

	for i := 0; i < s3.ParallelUpload; i++ {
		go s3.ToUpload()
	}
	for i := 0; i < s3.ParallelResize; i++ {
		go s3.ToResize()
	}

	return s3
}

func (s3 *Service) Upload(filePath, contentType, objectName string) error {
	ctx := context.Background()

	// Initialize minio client object.
	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return err
	}

	err = minioClient.MakeBucket(ctx, s3.bucketName, minio.MakeBucketOptions{Region: s3.location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, s3.bucketName)
		if errBucketExists != nil || !exists {
			return err
		}
	} else {
		logrus.Infof("S3: successfully created %s\n", s3.bucketName)
	}

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, s3.bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	logrus.Debugf("successfully uploaded %s of size %d\n", objectName, info.Size)

	return nil
}

func (s3 *Service) URL(objectName string) string {
	return fmt.Sprintf("https://%s.%s/%s", s3.bucketName, s3.endpoint, objectName)
}

func (s3 *Service) UploadPhoto(ctx context.Context, filePath string, userUUID uuid.UUID) (err error) {
	defer Span(NewSpan(ctx, "S3UploadPhoto"))()

	if ok, err := helpers.FileIsImage(filePath); !ok || err != nil {
		if err != nil {
			return err
		}
		return fmt.Errorf("file is not image")
	}

	wg := sync.WaitGroup{}
	wg.Add(len(s3.cropWidth))

	for _, width := range s3.cropWidth {
		s3.ToProcess(ToUpload{
			Ctx:          ctx,
			Path:         filePath,
			OriginalPath: filePath,
			UserUUID:     userUUID,
			Width:        width,
			Done:         &wg,
		})
	}

	if waitTimeout(&wg, s3.UploadTimeout) {
		return fmt.Errorf("upload photo timeout: (%s) (%s)", filePath, userUUID)
	}

	return nil
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func (s3 *Service) DeletePhoto(uid uuid.UUID) (err error) {
	ctx := context.Background()

	eg := errgroup.Group{}
	for _, size := range s3.cropWidth {
		func(size int) {
			eg.Go(func() error {
				objectName := s3.GetPhotoObjectName(uid, size)
				logrus.Debug("deleting photo: ", objectName)

				// Initialize minio client object.
				minioClient, err := minio.New(s3.endpoint, &minio.Options{
					Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
					Secure: s3.useSSL,
				})
				if err != nil {
					return err
				}

				err = minioClient.RemoveObject(ctx, s3.bucketName, objectName, minio.RemoveObjectOptions{
					ForceDelete: true,
				})
				if err != nil {
					logrus.Error(err)
					return err
				}

				return nil
			})
		}(size)
	}

	return eg.Wait()
}

func (s3 *Service) GetPhotoObjectName(userUUID uuid.UUID, size int) string {
	objectName := fmt.Sprintf("photos-%s.w%d.jpg", userUUID, size)
	if size == 0 {
		objectName = fmt.Sprintf("photos-%s.jpg", userUUID)
	}

	return objectName
}

func (s3 *Service) GetPhotoURL(userUUID uuid.UUID, size int) string {
	return s3.URL(s3.GetPhotoObjectName(userUUID, size))
}

func (s3 *Service) GetSmallPhoto(uid uuid.UUID) string {
	return s3.GetPhotoURL(uid, SmallPhotoSize)
}

func (s3 *Service) GetMediumPhoto(uid uuid.UUID) string {
	return s3.GetPhotoURL(uid, MediumPhotoSize)
}

func (s3 *Service) GetLargePhoto(uid uuid.UUID) string {
	return s3.GetPhotoURL(uid, LargePhotoSize)
}
