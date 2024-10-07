package agents

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
)

func New(repo *Repository) *Service {
	c := &Service{
		repo: repo,
	}

	return c
}

func (s *Service) Create(_ context.Context, a *domain.Agent) error {
	return s.repo.Create(a)
}

func (s *Service) Get(ctx context.Context, filter domain.AgentFilter) ([]domain.Agent, int64, error) {
	return s.repo.Get(ctx, filter)
}

func (s *Service) Delete(_ context.Context, uid uuid.UUID) error {
	return s.repo.Delete(uid)
}

func (s *Service) Update(_ context.Context, a *domain.Agent) error {
	return s.repo.Update(a)
}
