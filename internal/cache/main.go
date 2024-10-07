package cache

import (
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo *Repository

	cacheTaskTTL          int
	cacheTasksTTL         int
	cachePresignedURLsTTL int
}

func New(repo *Repository, opt *configs.Configs) *Service {
	if opt == nil {
		opt = &configs.Configs{}
		logrus.Warn("configs is nil, using default configs")
	}

	return &Service{
		repo:                  repo,
		cacheTaskTTL:          opt.CACHE_TASK,
		cacheTasksTTL:         opt.CACHE_TASKS,
		cachePresignedURLsTTL: opt.CACHE_PRE_SIGNED_URLS,
	}
}
