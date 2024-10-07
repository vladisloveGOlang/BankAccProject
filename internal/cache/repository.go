package cache

import (
	"github.com/krisch/crm-backend/pkg/redis"
)

type Repository struct {
	rds *redis.RDS
}

func NewRepository(rds *redis.RDS) *Repository {
	return &Repository{
		rds: rds,
	}
}
