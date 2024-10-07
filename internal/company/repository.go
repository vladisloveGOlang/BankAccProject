package company

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/cache"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	gorm         *postgres.GDB
	rds          *redis.RDS
	cacheService *cache.Service
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, cs *cache.Service) *Repository {
	return &Repository{
		gorm:         db,
		rds:          rds,
		cacheService: cs,
	}
}

func (r *Repository) PubUpdate() {
	err := r.rds.Publish(context.Background(), "update", "tag")
	logrus.Debug("pub update tag")
	if err != nil {
		logrus.Error(err)
	}
}

func (r *Repository) CreateTag(tag domain.Tag) (err error) {
	orm := &CompanyTags{}
	res := r.gorm.DB.Model(orm).
		Select("id").
		Where("name = ?", tag.Name).
		Where("company_uuid = ?", tag.CompanyUUID).
		Where("deleted_at is null").
		First(&orm)

	if res.RowsAffected > 0 {
		return errors.New("тег с таким именем уже существует")
	}

	orm = &CompanyTags{
		UUID:           tag.UUID,
		Name:           tag.Name,
		Color:          tag.Color,
		FederationUUID: tag.FederationUUID,
		CompanyUUID:    tag.CompanyUUID,
		CreatedByUUID:  tag.CreatedBy.UUID,
		CreatedBy:      tag.CreatedBy.Email,
	}

	err = r.gorm.DB.Create(&orm).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetTag(uid uuid.UUID) (orm CompanyTags, err error) {
	err = r.gorm.DB.Where("uid = ?", uid).First(&orm).Error
	return orm, err
}

func (r *Repository) GetTags(companyUUID uuid.UUID) (tags []CompanyTags, err error) {
	orm := []CompanyTags{}
	err = r.gorm.DB.
		Model(&orm).
		Where("company_uuid = ?", companyUUID).
		Where("deleted_at is null").
		Find(&orm).
		Error

	return orm, err
}

func (r *Repository) UpdateTag(uid uuid.UUID, name, color string) (err error) {
	res := r.gorm.DB.
		Model(&CompanyTags{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("name", name).
		Update("color", color).
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("нельзя обновлять удаленный тег")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteTag(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&CompanyTags{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("тег не найден или уже удален")
	}

	r.PubUpdate()

	// Remove tag from tasks
	uuids := []uuid.UUID{}
	err = r.gorm.DB.Raw("update tasks set tags = array_remove(tags, ?) where ? = ANY(tags) RETURNING uuid", uid.String(), uid.String()).Scan(&uuids).Error
	if res.Error != nil {
		return res.Error
	}

	for _, u := range uuids {
		r.cacheService.ClearTask(context.Background(), u)
	}

	return err
}

///

func (r *Repository) CreateCompanyPriority(cp domain.CompanyPriority) (err error) {
	orm := &CompanyPriority{}
	res := r.gorm.DB.Model(orm).
		Where("number = ?", cp.Number).
		Where("company_uuid = ?", cp.CompanyUUID).
		Where("deleted_at is null").
		First(&orm)

	if res.RowsAffected > 0 {
		return errors.New("приоритет с таким номером уже существует")
	}

	orm = &CompanyPriority{
		UUID:        cp.UUID,
		Name:        cp.Name,
		Number:      cp.Number,
		Color:       cp.Color,
		CompanyUUID: cp.CompanyUUID,
	}

	err = r.gorm.DB.Create(&orm).Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetCompanyPriority(uid uuid.UUID) (orm CompanyPriority, err error) {
	err = r.gorm.DB.
		Where("uid = ?", uid).
		Where("deleted_at is null").
		First(&orm).
		Error
	return orm, err
}

func (r *Repository) GetCompanyPriorities(companyUUID uuid.UUID) (tags []CompanyPriority, err error) {
	orm := []CompanyPriority{}
	err = r.gorm.DB.
		Model(&orm).
		Where("company_uuid = ?", companyUUID).
		Where("deleted_at is null").
		Find(&orm).
		Error

	return orm, err
}

func (r *Repository) UpdateCompanyPriority(uid uuid.UUID, name, color string) (err error) {
	res := r.gorm.DB.
		Model(&CompanyPriority{}).
		Where("uuid = ?", uid).
		Update("name", name).
		Update("color", color).
		Where("deleted_at is null").
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("приоритет удален или не найден")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteCompanyPriority(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&CompanyPriority{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("тег не найден или уже удален")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) UpdateSmsOptions(uid uuid.UUID, so SmsOptions) error {
	return r.gorm.DB.
		Model(&Company{
			UUID: uid,
		}).
		Update("sms_options", so).
		Error
}

func (r *Repository) GetSmsOptions(uid uuid.UUID) (orm Company, err error) {
	err = r.gorm.DB.
		Select("uuid", "sms_options").
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		First(&orm).
		Error
	return orm, err
}
