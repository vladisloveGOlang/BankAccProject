package notifications

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/aggregates"
	"github.com/krisch/crm-backend/internal/dictionary"
)

type Service struct {
	repo *Repository
	dict *dictionary.Service
	aggs *aggregates.Service
}

func New(repo *Repository, aggs *aggregates.Service, dict *dictionary.Service) *Service {
	return &Service{
		repo: repo,
		aggs: aggs,
		dict: dict,
	}
}

func (s *Service) GetNotification(email string) ([]dto.NotificationDTO, error) {
	dtos, err := s.repo.GetNotification(email)

	return dtos, err
}

func (s *Service) GetNotificationCount(email, kind string, uid uuid.UUID) (map[string]string, error) {
	return s.repo.GetNotificationCount(email, kind, uid)
}

func (s *Service) RemoveNotification(email, kind string, uid uuid.UUID) error {
	err := s.repo.RemoveNotification(email, kind+":"+uid.String())

	return err
}

func (s *Service) RemoveNotifications(ctx context.Context, email string) error {
	err := s.repo.RemoveNotifications(ctx, email)

	return err
}

func (s *Service) Count(ctx context.Context, email string) (int64, error) {
	defer Span(NewSpan(ctx, "NotificationsCount"))()
	return s.repo.Count(email)
}

func (s *Service) ToggleStarNotification(ctx context.Context, typeName, email string, uid uuid.UUID, star bool) error {
	defer Span(NewSpan(ctx, "ToggleStarNotification"))()
	return s.repo.ToggleStarNotification(email, typeName+":"+uid.String(), star)
}

func (s *Service) HideNotification(ctx context.Context, typeName, email string, uid uuid.UUID) error {
	defer Span(NewSpan(ctx, "HideNotification"))()
	return s.repo.HideNotification(email, typeName+":"+uid.String())
}
