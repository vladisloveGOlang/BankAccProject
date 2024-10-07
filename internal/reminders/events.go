package reminders

import "github.com/google/uuid"

func (s *Service) OnReminderWasUpdatedOrCreated(fn func(uuid.UUID, uuid.UUID, []string) error) {
	s.onReminderWasUpdatedOrCreated = fn
}
