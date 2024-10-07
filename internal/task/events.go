package task

import "github.com/google/uuid"

func (s *Service) OnTaskUpdatedOrCreated(fn func(uuid.UUID, []string) error) {
	s.onTaskUpdatedOrCreated = fn
}

func (s *Service) OnOpenTask(fn func(uuid.UUID, string) error) {
	s.onOpenTask = fn
}
