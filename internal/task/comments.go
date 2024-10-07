package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

func (s *Service) CreateComment(ctx context.Context, uid uuid.UUID, cm domain.Comment) (err error) {
	task, err := s.GetTask(ctx, uid, []string{})
	if err != nil {
		return err
	}

	err = s.commentService.CreateComment(ctx, cm)
	if err != nil {
		return err
	}

	notify := lo.Filter(task.People, func(email string, _ int) bool {
		return email != cm.CreatedBy
	})

	err = s.TaskWasUpdatedOrCreated(uid, notify)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateComment(ctx context.Context, uid uuid.UUID, cm domain.Comment) (err error) {
	task, err := s.GetTask(ctx, uid, []string{})
	if err != nil {
		return err
	}

	err = s.commentService.UpdateComment(ctx, cm)
	if err != nil {
		return err
	}

	notify := lo.Filter(task.People, func(email string, _ int) bool {
		return email != cm.CreatedBy
	})

	err = s.TaskWasUpdatedOrCreated(uid, notify)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteComment(ctx context.Context, taskUID, comentUID uuid.UUID, deletedBy string) (err error) {
	task, err := s.GetTask(ctx, taskUID, []string{})
	if err != nil {
		return err
	}

	err = s.commentService.DeleteComment(ctx, task.UUID, comentUID, nil)
	if err != nil {
		return err
	}

	notify := lo.Filter(task.People, func(email string, _ int) bool {
		return email != deletedBy
	})

	err = s.TaskWasUpdatedOrCreated(taskUID, notify)
	if err != nil {
		return err
	}

	return nil
}
