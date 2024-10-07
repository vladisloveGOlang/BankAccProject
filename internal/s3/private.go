package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/cache"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type ServicePrivate struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	location        string // "ru-central1"
	useSSL          bool
	publicURL       string
	backendURL      string

	repo  *Repository
	cache *cache.Service
}

type ConfPrivate struct {
	Endpoint        string
	BackendURL      string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Location        string
	UseSSL          bool
	PublicURL       string
}

func NewPrivate(conf ConfPrivate, repo *Repository, cs *cache.Service) *ServicePrivate {
	s3 := &ServicePrivate{
		repo:  repo,
		cache: cs,

		endpoint:        conf.Endpoint,
		accessKeyID:     conf.AccessKeyID,
		secretAccessKey: conf.SecretAccessKey,
		bucketName:      conf.BucketName,
		location:        conf.Location,
		useSSL:          conf.UseSSL,
		publicURL:       conf.PublicURL,
		backendURL:      conf.BackendURL,
	}

	return s3
}

func (s3 *ServicePrivate) UploadTaskFile(federatonUUID, taskUUID uuid.UUID, fileName, filePath string, userUUID uuid.UUID) (file File, err error) {
	ext := helpers.FileExt(filePath)
	objectName := fmt.Sprintf("%s/task/%s/%s%s", federatonUUID, taskUUID, uuid.New().String(), ext)

	fileDTO, err := NewFileDTO(fileName, filePath, objectName, userUUID)
	if err != nil {
		return file, err
	}

	file = File{
		UUID: uuid.New(),

		Type:     "task",
		TypeUUID: taskUUID,

		Name:       fileDTO.Name,
		ObjectName: objectName,
		Size:       fileDTO.Size,
		Ext:        fileDTO.Ext,

		ImgWidth:  fileDTO.Width,
		ImgHeight: fileDTO.Height,

		MimeType:   fileDTO.ContentType,
		BucketName: s3.bucketName,
		Endpoint:   s3.endpoint,
		CreatedBy:  userUUID,
	}

	return s3.uploadFile(file, filePath)
}

func (s3 *ServicePrivate) UploadTaskCommentFile(federatonUUID, taskUUID, commentUUID uuid.UUID, fileName, filePath string, userUUID uuid.UUID) (file File, err error) {
	ext := helpers.FileExt(filePath)
	objectName := fmt.Sprintf("%s/task/%s/%s%s", federatonUUID, taskUUID, uuid.New().String(), ext)

	fileDTO, err := NewFileDTO(fileName, filePath, objectName, userUUID)
	if err != nil {
		return file, err
	}

	file = File{
		UUID: uuid.New(),

		Type:     "comment",
		TypeUUID: commentUUID,

		Name:       fileDTO.Name,
		ObjectName: objectName,
		Size:       fileDTO.Size,
		Ext:        fileDTO.Ext,

		ImgWidth:  fileDTO.Width,
		ImgHeight: fileDTO.Height,

		MimeType:   fileDTO.ContentType,
		BucketName: s3.bucketName,
		Endpoint:   s3.endpoint,
		CreatedBy:  userUUID,
	}

	return s3.uploadFile(file, filePath)
}

func (s3 *ServicePrivate) uploadFile(file File, filePath string) (File, error) {
	err := s3.repo.Create(file)
	if err != nil {
		return file, err
	}

	ctx := context.Background()

	// Initialize minio client object.
	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return file, err
	}

	err = minioClient.MakeBucket(ctx, file.BucketName, minio.MakeBucketOptions{Region: s3.location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, s3.bucketName)
		if errBucketExists != nil || !exists {
			return file, err
		}
	} else {
		logrus.Infof("S3: successfully created %s\n", file.BucketName)
	}

	info, err := minioClient.FPutObject(ctx, file.BucketName, file.ObjectName, filePath, minio.PutObjectOptions{ContentType: file.MimeType})
	if err != nil {
		return file, err
	}

	logrus.Debugf("successfully uploaded %s of size %d\n", file.ObjectName, info.Size)

	return file, err
}

func (s3 *ServicePrivate) DeleteFile(file File) error {
	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return fmt.Errorf("S3: %w", err)
	}

	ctx := context.Background()

	err = minioClient.RemoveObject(ctx, file.BucketName, file.ObjectName, minio.RemoveObjectOptions{
		ForceDelete: true,
	})

	if err != nil {
		return fmt.Errorf("S3: %w", err)
	}

	return err
}

func (s3 *ServicePrivate) Delete(fileUUID uuid.UUID) error {
	file, err := s3.repo.GetFile(fileUUID)
	if err != nil {
		return err
	}

	err = s3.repo.MarkForDelete(fileUUID)
	if err != nil {
		return err
	}

	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return fmt.Errorf("S3: %w", err)
	}

	ctx := context.Background()

	err = minioClient.RemoveObject(ctx, file.BucketName, file.ObjectName, minio.RemoveObjectOptions{
		ForceDelete: true,
	})

	if err != nil {
		return fmt.Errorf("S3: %w", err)
	}

	err = s3.repo.Delete(fileUUID)
	if err != nil {
		return err
	}

	logrus.Debugf("successfully deleted")

	return err
}

func (s3 *ServicePrivate) Rename(fileUUID uuid.UUID, name string) error {
	err := s3.repo.Rename(fileUUID, name)
	if err != nil {
		return err
	}

	return err
}

func (s3 *ServicePrivate) PresignedURL(name, objectName string) (res string, err error) {
	// Initialize minio client object.
	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return res, err
	}

	ctx := context.Background()

	reqParams := make(url.Values)

	reqParams.Set("response-content-disposition", fmt.Sprintf("filename=\"%q\"", name))

	presignedURL, err := minioClient.PresignedGetObject(ctx, s3.bucketName, objectName, time.Second*24*60*60, reqParams)
	if err != nil {
		return res, err
	}

	return presignedURL.String(), err
}

func (s3 *ServicePrivate) PresignedURLFromFile(fileUUID uuid.UUID) (res string, err error) {
	file, err := s3.repo.GetFile(fileUUID)
	if err != nil {
		return res, err
	}

	return s3.PresignedURL(file.Name, file.ObjectName)
}

func (s3 *ServicePrivate) GetTaskFiles(taskUUID uuid.UUID, openImages bool) (dmns []domain.File, err error) {
	files, err := s3.repo.GetTaskFiles(taskUUID)
	if err != nil {
		return dmns, err
	}

	return lo.Map(files, func(item File, index int) domain.File {
		fileURL := fmt.Sprintf("%s/task/%s/upload/%s", s3.backendURL, item.TypeUUID, item.UUID)

		if openImages && helpers.FileMimeToPreview(item.MimeType) {
			urlFromRedis, err := s3.cache.GetURL(context.Background(), item.UUID)
			if err == nil && urlFromRedis != "" {
				fileURL = urlFromRedis
			} else {
				presignedURL, err := s3.PresignedURL(item.Name, item.ObjectName)
				if err != nil {
					logrus.Warn(err)
				} else {
					fileURL = presignedURL
				}

				s3.cache.CacheURL(context.Background(), item.UUID, fileURL)
			}
		}

		return domain.File{
			UUID:      item.UUID,
			Name:      item.Name,
			Ext:       item.Ext,
			Size:      item.Size,
			URL:       fileURL,
			CreatedAt: item.CreatedAt,
			CreatedBy: item.CreatedBy,
		}
	}), err
}

func (s3 *ServicePrivate) GetCommentFiles(uid uuid.UUID, openImages bool) (dmns []domain.File, err error) {
	files, err := s3.repo.GetCommentFiles(uid)
	if err != nil {
		return dmns, err
	}

	return lo.Map(files, func(item File, index int) domain.File {
		fileURL := fmt.Sprintf("%s/task/%s/upload/%s", s3.backendURL, item.TypeUUID, item.UUID)

		if openImages && helpers.FileMimeToPreview(item.MimeType) {
			presignedURL, err := s3.PresignedURL(item.Name, item.ObjectName)
			if err != nil {
				logrus.Warn(err)
			} else {
				fileURL = presignedURL
			}
		}

		return domain.File{
			UUID: item.UUID,
			Name: item.Name,
			Ext:  item.Ext,
			Size: item.Size,
			URL:  fileURL,
		}
	}), err
}

// @todo: in poc.
func (s3 *ServicePrivate) DangerousWipeS3FederationData(existFederations []domain.Federation) (uids []string, err error) {
	minioClient, err := minio.New(s3.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.accessKeyID, s3.secretAccessKey, ""),
		Secure: s3.useSSL,
	})
	if err != nil {
		return []string{}, fmt.Errorf("S3: %w", err)
	}

	ctx := context.Background()

	deletedTotal := 0
	for i := range minioClient.ListObjects(ctx, s3.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	}) {
		if deletedTotal > 1000 {
			return []string{}, nil
		}

		hasPrefix := false
		for _, federation := range existFederations {
			if strings.HasPrefix(i.Key, federation.UUID.String()+"/") {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			deletedTotal++
			logrus.WithField("key", i.Key).Info("deleting federation s3 data")

			err = minioClient.RemoveObject(ctx, s3.bucketName, i.Key, minio.RemoveObjectOptions{
				ForceDelete: true,
			})
			if err != nil {
				logrus.Error(err)
				return []string{}, fmt.Errorf("S3: %w", err)
			}
		}

		logrus.WithField("total", deletedTotal).Info("successfully deleted federation s3 data")
	}

	return []string{}, err
}
