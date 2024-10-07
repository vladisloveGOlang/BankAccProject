package sms

import (
	"context"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	gorm *postgres.GDB
}

func NewRepository(db *postgres.GDB) *Repository {
	return &Repository{
		gorm: db,
	}
}

func (r *Repository) Create(s *domain.Sms) error {
	return r.gorm.DB.Create(&sms{
		UUID:           s.UUID,
		FederationUUID: s.FederationUUID,
		CompanyUUID:    s.CompanyUUID,
		CreatedBy:      s.CreatedBy,
		CreatedByUUID:  s.CreatedByUUID,
		To:             s.To,
		Text:           s.Text,
		From:           s.From,
	}).Error
}

func (r *Repository) GetSms(_ context.Context, filter dto.SmsFilterDTO) (dms []domain.Sms, total int64, err error) {
	orms := []sms{}

	query := r.gorm.DB

	query = query.Order("created_at desc")

	if filter.CompanyUUID != nil {
		query = query.Where("company_uuid = ?", *filter.CompanyUUID)
	}

	if filter.IsMy != nil && *filter.IsMy && filter.MyEmail != nil {
		query = query.Where("created_by = ?", filter.MyEmail)
	}

	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	} else {
		query = query.Limit(5)
	}

	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	} else {
		query = query.Offset(0)
	}

	query = query.Where("deleted_at is null")

	query = query.Select("*, count(*) OVER() AS total")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&orms)
	})

	logrus.Debug("sql: ", sql)

	result := query.Find(&orms)

	if result.Error != nil {
		return dms, -1, result.Error
	}

	if len(orms) > 0 {
		total = orms[0].Total
	}

	dms = helpers.Map(orms, func(item sms, i int) domain.Sms {
		return domain.Sms{
			UUID:           item.UUID,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,

			CreatedBy:     item.CreatedBy,
			CreatedByUUID: item.CreatedByUUID,

			To:   item.To,
			From: item.From,
			Text: item.Text,

			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	})

	return dms, total, nil
}
