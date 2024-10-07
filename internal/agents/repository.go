package agents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/samber/lo"
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

func (r *Repository) Create(s *domain.Agent) error {
	return r.gorm.DB.Create(&Agent{
		UUID:           s.UUID,
		FederationUUID: s.FederationUUID,
		CompanyUUID:    s.CompanyUUID,
		CreatedBy:      s.CreatedBy,
		CreatedByUUID:  s.CreatedByUUID,

		Name: s.Name,
		Contacts: lo.Map(s.Contacts, func(c domain.AgentContacts, _ int) Contacts {
			return Contacts{
				Type: c.Type,
				Val:  c.Val,
			}
		}),
	}).Error
}

func (r *Repository) Get(_ context.Context, filter domain.AgentFilter) (dms []domain.Agent, total int64, err error) {
	if filter.FederationUUID == uuid.Nil {
		return nil, -1, errors.New("federation uuid is required")
	}

	orms := []Agent{}

	query := r.gorm.DB

	query = query.Order("created_at desc")

	query = query.Where("federation_uuid = ?", filter.FederationUUID)

	if filter.CompanyUUID != nil {
		query = query.Where("company_uuid = ?", *filter.CompanyUUID)
	}

	if filter.Name != nil {
		query = query.Where("name ilike ?", *filter.Name+"%")
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

	dms = helpers.Map(orms, func(item Agent, i int) domain.Agent {
		return domain.Agent{
			UUID:           item.UUID,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,

			CreatedBy:     item.CreatedBy,
			CreatedByUUID: item.CreatedByUUID,

			Name: item.Name,
			Contacts: lo.Map(item.Contacts, func(c Contacts, _ int) domain.AgentContacts {
				return domain.AgentContacts{
					Type: c.Type,
					Val:  c.Val,
				}
			}),

			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
		}
	})

	return dms, total, nil
}

func (r *Repository) Update(s *domain.Agent) error {
	return r.gorm.DB.Model(&Agent{}).
		Where("uuid = ?", s.UUID).
		Where("deleted_at is null").
		Updates(&Agent{
			Name: s.Name,
			Contacts: lo.Map(s.Contacts, func(c domain.AgentContacts, _ int) Contacts {
				return Contacts{
					Type: c.Type,
					Val:  c.Val,
				}
			}),
		}).Error
}

func (r *Repository) Delete(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&Agent{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("агент не найдена")
	}

	return res.Error
}
