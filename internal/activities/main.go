package activities

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/samber/lo"
)

type (
	Meta map[string]interface{}
)

type Service struct {
	repo *Repository
	dict *dictionary.Service
}

type ActivityMeta struct {
	Old  interface{} `json:"old"`
	New  interface{} `json:"new"`
	Name string      `json:"name"`
}

func New(repo *Repository, dict *dictionary.Service) *Service {
	s := &Service{
		repo: repo,
		dict: dict,
	}

	return s
}

func (s *Service) CreateActivity(activity *Activity) error {
	return s.repo.CreateActivity(activity)
}

func (s *Service) GetTaskActivities(taskUID uuid.UUID, limit, offset int) ([]domain.Activity, int64, error) {
	orms, total, err := s.repo.GetTaskActivities(taskUID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return lo.Map(orms, func(orm Activity, _ int) domain.Activity {
		return domain.Activity{
			UUID:        orm.UUID,
			EntityUUID:  orm.EntityUUID,
			EntityType:  orm.EntityType,
			Description: orm.Description,
			CreatedBy: domain.User{
				UUID:  orm.CreatedByUUID,
				Email: orm.CreatedBy,
			},
			CreatedAt: orm.CreatedAt,
			Meta:      orm.Meta,
			Type:      int(orm.Type),
		}
	}), total, nil
}
