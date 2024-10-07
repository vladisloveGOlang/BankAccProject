package gates

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

type Repository struct {
	gorm *postgres.GDB
	rds  *redis.RDS
}

func NewRepository(db *postgres.GDB, rds *redis.RDS) *Repository {
	return &Repository{
		gorm: db,
		rds:  rds,
	}
}

func (r *Repository) PubUpdate() {
	err := r.rds.Publish(context.Background(), "update", "acl")
	logrus.Debug("pub update acl")
	if err != nil {
		logrus.Error(err)
	}
}

func (r *Repository) CreateOrUpdatePermisson(perm *domain.Permission) (err error) {
	orm := &Permission{
		UserUUID:       perm.UserUUID,
		FederationUUID: perm.FederationUUID,
		DeletedAt:      nil,
		Rules:          perm.Rules,
	}

	err = r.gorm.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_uuid"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"rules":      perm.Rules,
			"updated_at": "now()",
			"deleted_at": nil,
		}),
	}).Create(&orm).Error

	perm.UpdatedAt = orm.UpdatedAt

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeletePermission(userUUID uuid.UUID) error {
	res := r.gorm.DB.
		Model(&Permission{}).
		Where("user_uuid = ?", userUUID).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("правило не найдено")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return res.Error
}

func (r *Repository) GetPermisson(userUUID uuid.UUID) (dm domain.Permission, err error) {
	orm := &Permission{}

	res := r.gorm.DB.Model(&orm).
		Where("user_uuid = ?", userUUID).
		Where("deleted_at is null").
		Find(&orm)

	if res.RowsAffected == 0 {
		return dm, dto.NotFoundErr("permission not found")
	}

	return domain.Permission{
		UserUUID:       orm.UserUUID,
		FederationUUID: orm.FederationUUID,
		Rules:          orm.Rules,
		CreatedAt:      orm.CreatedAt,
		UpdatedAt:      orm.UpdatedAt,
	}, nil
}
