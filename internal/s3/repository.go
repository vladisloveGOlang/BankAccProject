package s3

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/pkg/postgres"
)

type Repository struct {
	gorm *postgres.GDB
}

func NewRepository(db *postgres.GDB) *Repository {
	return &Repository{
		gorm: db,
	}
}

func (r *Repository) Create(orm File) error {
	err := r.gorm.DB.Create(&orm).Error
	if err != nil {
		return err
	}

	return err
}

func (r *Repository) GetFile(fileUUID uuid.UUID) (file File, err error) {
	res := r.gorm.DB.
		Model(&File{}).
		Where("uuid = ?", fileUUID).
		Where("deleted_at IS NULL").
		First(&file)

	if res.RowsAffected == 0 {
		return file, dto.NotFoundErr("файл не найден")
	}

	return file, err
}

func (r *Repository) GetTaskFiles(taskUUID uuid.UUID) (files []File, err error) {
	res := r.gorm.DB.
		Model(&File{}).
		Where("type = ?", "task").
		Where("type_uuid = ?", taskUUID).
		Where("deleted_at IS NULL").
		Find(&files)

	return files, res.Error
}

func (r *Repository) GetCommentFiles(commnetUUID uuid.UUID) (files []File, err error) {
	res := r.gorm.DB.
		Model(&File{}).
		Where("type = ?", "comment").
		Where("type_uuid = ?", commnetUUID).
		Where("deleted_at IS NULL").
		Find(&files)

	return files, res.Error
}

func (r *Repository) GetCommentsFiles(commnetsUUID []uuid.UUID) (files []File, err error) {
	res := r.gorm.DB.
		Model(&File{}).
		Where("type = ?", "comment").
		Where("type_uuid = IN", commnetsUUID).
		Where("deleted_at IS NULL").
		Find(&files)

	return files, res.Error
}

func (r *Repository) MarkForDelete(fileUUID uuid.UUID) error {
	res := r.gorm.DB.
		Model(&File{}).
		Where("uuid = ?", fileUUID).
		UpdateColumn("to_deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("файл не найден")
	}

	return res.Error
}

func (r *Repository) Delete(fileUUID uuid.UUID) error {
	res := r.gorm.DB.
		Model(&File{}).
		Where("uuid = ?", fileUUID).
		Where("deleted_at IS NULL").
		UpdateColumn("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("файл не найден")
	}

	return res.Error
}

func (r *Repository) Rename(fileUUID uuid.UUID, name string) error {
	res := r.gorm.DB.
		Model(&File{}).
		Where("uuid = ?", fileUUID).
		Where("deleted_at IS NULL").
		UpdateColumn("name", name)

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("файл не найден")
	}

	return res.Error
}
