package dictionary

import (
	"time"

	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/prometheus/client_golang/prometheus"
)

type Repository struct {
	gorm      *postgres.GDB
	rds       *redis.RDS
	histogram *prometheus.HistogramVec
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, metrics *helpers.MetricsCounters) *Repository {
	return &Repository{
		gorm:      db,
		rds:       rds,
		histogram: metrics.RepoHistogram,
	}
}

func (r *Repository) FetchUsers(updatedAt time.Time) (items []User, err error) {
	err = r.gorm.DB.Table("users").
		Select("users.*, fu.federation_uuid as federation_uuid").
		Joins("LEFT JOIN federation_users fu ON fu.user_uuid = users.uuid").
		Where("users.updated_at >= ? or fu.updated_at >= ?", updatedAt, updatedAt).
		Find(&items).
		Error

	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchCompanyTags(updatedAt time.Time) (items []CompanyTags, err error) {
	err = r.gorm.DB.Table("company_tags").
		Where("updated_at >= ?", updatedAt).
		Find(&items).
		Error

	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchFederations(updatedAt time.Time) (items []Federation, err error) {
	err = r.gorm.DB.Table("federations").Where("updated_at >= ?", updatedAt).Find(&items).Error
	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchCompanies(updatedAt time.Time) (items []Company, err error) {
	err = r.gorm.DB.Table("companies").Where("updated_at >= ?", updatedAt).Find(&items).Error
	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchProjects(updatedAt time.Time) (items []Project, err error) {
	err = r.gorm.DB.Table("projects").Where("updated_at >= ?", updatedAt).Find(&items).Error
	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchCompanyFields(updatedAt time.Time) (items []CompanyFields, err error) {
	err = r.gorm.DB.Table("company_fields").
		Select("company_fields.*, coalesce( json_agg(pf.project_uuid) FILTER (WHERE pf.project_uuid is not null), '[]' ) as projects_uuid").
		Joins("LEFT JOIN project_fields pf ON pf.company_field_uuid = company_fields.uuid").
		Where("company_fields.updated_at >= ? or pf.updated_at >= ?", updatedAt, updatedAt).
		Group("company_fields.uuid").
		Find(&items).Error
	if err != nil {
		return items, err
	}
	return items, nil
}

func (r *Repository) FetchProjectFields(updatedAt time.Time) (items []ProjectFields, err error) {
	err = r.gorm.DB.Table("project_fields").
		Select("project_fields.project_uuid project_uuid, project_fields.uuid uuid, project_fields.company_uuid company_uuid, cf.name name, cf.hash hash, project_fields.style, project_fields.required_on_statuses required_on_statuses").
		Joins("LEFT JOIN company_fields cf ON project_fields.company_field_uuid = cf.uuid").
		Where("project_fields.updated_at >= ? or cf.updated_at >= ?", updatedAt, updatedAt).
		Find(&items).Error
	if err != nil {
		return items, err
	}
	return items, nil
}

func (r *Repository) FetchCatalogFields(updatedAt time.Time) (items []CatalogFields, err error) {
	err = r.gorm.DB.Table("catalog_fields").Where("updated_at >= ?", updatedAt).Find(&items).Error
	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchUserFederation(updatedAt time.Time) (items []UserFederation, err error) {
	err = r.gorm.DB.Table("federation_users").
		Where("updated_at >= ?", updatedAt).
		Find(&items).
		Error

	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchUserCompany(updatedAt time.Time) (items []UsersCompanies, err error) {
	err = r.gorm.DB.Table("company_users").
		Where("updated_at >= ?", updatedAt).
		Find(&items).
		Error

	if err != nil {
		return items, err
	}

	return items, nil
}

func (r *Repository) FetchCompanyPriority(updatedAt time.Time) (items []CompanyPriority, err error) {
	err = r.gorm.DB.Table("company_priorities").
		Where("updated_at >= ?", updatedAt).
		Find(&items).
		Error

	if err != nil {
		return items, err
	}

	return items, nil
}
