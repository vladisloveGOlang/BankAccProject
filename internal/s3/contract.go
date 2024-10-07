package s3

import "github.com/google/uuid"

type IS3 interface {
	StoreUserPhoto(path, userUUID uuid.UUID) error
	DeleteUserPhoto(userUUID uuid.UUID) error
	GetUserPhotoUrl(userUUID uuid.UUID, size int) string

	StoreTaskFile(path string, userUUID uuid.UUID) error
}
