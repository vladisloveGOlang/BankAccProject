package reminders

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/samber/lo"
	"gorm.io/gorm/clause"
)

type Repository struct {
	gorm *postgres.GDB
}

func NewRepository(db *postgres.GDB) *Repository {
	return &Repository{
		gorm: db,
	}
}

func (r *Repository) Create(dm domain.Reminder) (err error) {
	orm := &Reminder{
		UUID:          dm.UUID,
		CreatedBy:     dm.CreatedBy,
		CreatedByUUID: dm.CreatedByUUID,
		TaskUUID:      dm.TaskUUID,
		DateFrom:      dm.DateFrom,
		DateTo:        dm.DateTo,
		Description:   dm.Description,
		Type:          dm.Type,
		UserUUID:      dm.UserUUID,
	}

	err = r.gorm.DB.Create(&orm).Error

	return err
}

func (r *Repository) Put(dm domain.Reminder) (err error) {
	orm := &Reminder{
		UUID:        dm.UUID,
		DateFrom:    dm.DateFrom,
		DateTo:      dm.DateTo,
		Description: dm.Description,
		Comment:     dm.Comment,
		Type:        dm.Type,
		UserUUID:    dm.UserUUID,
	}

	res := r.gorm.DB.Save(&orm)

	if res.Error != nil {
		return res.Error
	}

	return res.Error
}

func (r *Repository) DeleteByUUID(uid uuid.UUID) error {
	orm := &Reminder{}

	res := r.gorm.DB.
		Model(orm).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "task_uuid"}}}).
		Where("uuid = ?", uid).
		Where("deleted_at IS NULL").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("дело не найдено")
	}

	return res.Error
}

func (r *Repository) GetByUser(uid uuid.UUID) (dms []domain.Reminder, err error) {
	orm := []Reminder{}

	err = r.gorm.DB.
		Where("(created_by_uuid = ? or user_uuid = ?)", uid, uid).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	dms = lo.Map(orm, func(item Reminder, i int) domain.Reminder {
		return domain.Reminder{
			UUID:          item.UUID,
			CreatedBy:     item.CreatedBy,
			CreatedByUUID: item.CreatedByUUID,
			TaskUUID:      item.TaskUUID,
			DateFrom:      item.DateFrom,
			DateTo:        item.DateTo,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
			Description:   item.Description,
			Comment:       item.Comment,
			Type:          item.Type,
			UserUUID:      item.UserUUID,
			Status:        item.Status,
		}
	})

	return dms, nil
}

func (r *Repository) ChangeField(uid uuid.UUID, fieldName string, value interface{}) error {
	res := r.gorm.DB.
		Model(&Reminder{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update(fieldName, value).
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("нельзя обновлять удаленное дело")
	}

	return res.Error
}

func (r *Repository) GetByTask(taskUUID uuid.UUID) (dms []domain.Reminder, err error) {
	orm := []Reminder{}

	err = r.gorm.DB.
		Where("task_uuid = ?", taskUUID).
		Where("deleted_at IS NULL").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	dms = lo.Map(orm, func(item Reminder, i int) domain.Reminder {
		return domain.Reminder{
			UUID:          item.UUID,
			CreatedBy:     item.CreatedBy,
			CreatedByUUID: item.CreatedByUUID,
			TaskUUID:      item.TaskUUID,
			UserUUID:      item.UserUUID,
			DateFrom:      item.DateFrom,
			DateTo:        item.DateTo,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
			Description:   item.Description,
			Comment:       item.Comment,
			Type:          item.Type,
		}
	})

	return dms, nil
}

func (r *Repository) Get(uid uuid.UUID) (dms domain.Reminder, err error) {
	orm := Reminder{}

	err = r.gorm.DB.
		Where("uuid = ?", uid).
		Where("deleted_at IS NULL").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	return domain.Reminder{
		UUID:          orm.UUID,
		CreatedBy:     orm.CreatedBy,
		CreatedByUUID: orm.CreatedByUUID,
		TaskUUID:      orm.TaskUUID,
		UserUUID:      orm.UserUUID,
		DateFrom:      orm.DateFrom,
		DateTo:        orm.DateTo,
		CreatedAt:     orm.CreatedAt,
		UpdatedAt:     orm.UpdatedAt,
		Description:   orm.Description,
	}, nil
}

func (r *Repository) GetRemindersNames(_ context.Context, uids []uuid.UUID) (withName []domain.Reminder, err error) {
	err = r.gorm.DB.
		Model(&Reminder{}).
		Select("uuid, description").
		Where("uuid in ?", uids).
		Where("deleted_at is null").
		Find(&withName).
		Error

	if err != nil {
		return withName, err
	}

	return withName, nil
}
