package health

import (
	"context"
	"errors"

	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
)

type Service struct {
	pdb *postgres.GDB
	rdb *redis.RDS
}

func NewHealthService(pdb *postgres.GDB, rdb *redis.RDS) *Service {
	return &Service{
		pdb: pdb,
		rdb: rdb,
	}
}

func (s *Service) PingPostgres() (bool, error) {
	if s.pdb == nil {
		return false, errors.New("нет соединения")
	}

	err := s.pdb.DB.Select("select 1").Error
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) PingRedis() (bool, error) {
	if s.rdb == nil {
		return false, errors.New("нет соединения")
	}

	err := s.rdb.Ping(context.TODO())
	if err != nil {
		return false, err
	}
	return true, nil
}
