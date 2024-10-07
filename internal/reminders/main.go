package reminders

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo *Repository
	dict *dictionary.Service

	onReminderWasUpdatedOrCreated func(uuid.UUID, uuid.UUID, []string) error
}

func New(repo *Repository, dict *dictionary.Service) *Service {
	// @todo: rm task service mv to cache service
	return &Service{
		repo: repo,
		dict: dict,
	}
}

func (s *Service) ReminderWasUpdatedOrCreated(uid, taskUUID uuid.UUID, people []string) error {
	if s.onReminderWasUpdatedOrCreated != nil {
		return s.onReminderWasUpdatedOrCreated(uid, taskUUID, people)
	}

	logrus.Error("onTaskUpdatedOrCreated is nil")

	return nil
}

func (s *Service) Create(r domain.Reminder) (err error) {
	if r.DateFrom != nil && r.DateTo != nil && !helpers.IsTheSameDay(*r.DateFrom, *r.DateTo) {
		return fmt.Errorf("даты должны быть в один день")
	}

	err = s.repo.Create(r)

	if r.UserUUID != nil {
		cuser, cu := s.dict.FindUserByUUID(*r.UserUUID)
		if !cu {
			return fmt.Errorf("reminder user not found by uuid: %s", r.UserUUID.String())
		}

		err := s.ReminderWasUpdatedOrCreated(r.UUID, r.TaskUUID, []string{cuser.Email})
		if err != nil {
			logrus.WithError(err).Error("ReminderWasUpdatedOrCreated error")
		}
	}

	return err
}

func (s *Service) Put(userEmail string, r domain.Reminder) (err error) {
	if r.DateFrom != nil && r.DateTo != nil && !helpers.IsTheSameDay(*r.DateFrom, *r.DateTo) {
		return fmt.Errorf("даты должны быть в один день")
	}

	err = s.repo.Put(r)
	if err == nil {
		people, err := s.GetPeople(r)
		people = lo.Filter(people, func(email string, _ int) bool {
			return email != userEmail
		})

		if err != nil {
			logrus.WithError(err).Error("GetPeople error")
			return err
		}

		err = s.ReminderWasUpdatedOrCreated(r.UUID, r.TaskUUID, people)
		if err != nil {
			logrus.WithError(err).Error("ReminderWasUpdatedOrCreated error")
		}
	}

	return err
}

func (s *Service) PatchStatus(userEmail string, r domain.Reminder, status int) (err error) {
	err = s.repo.ChangeField(r.UUID, "status", status)

	if err == nil {
		people, err := s.GetPeople(r)
		people = lo.Filter(people, func(email string, _ int) bool {
			return email != userEmail
		})

		if err != nil {
			logrus.WithError(err).Error("GetPeople error")
			return err
		}

		err = s.ReminderWasUpdatedOrCreated(r.UUID, r.TaskUUID, people)
		if err != nil {
			logrus.WithError(err).Error("ReminderWasUpdatedOrCreated error")
		}
	}

	return err
}

func (s *Service) DeleteByUUID(uid uuid.UUID) (err error) {
	r, err := s.Get(uid)
	if err != nil {
		return err
	}

	err = s.repo.DeleteByUUID(r.UUID)

	if err == nil {
		people, err := s.GetPeople(r)
		if err != nil {
			logrus.WithError(err).Error("GetPeople error")
			return err
		}

		err = s.ReminderWasUpdatedOrCreated(r.UUID, r.TaskUUID, people)
		if err != nil {
			logrus.WithError(err).Error("ReminderWasUpdatedOrCreated error")
		}
	}

	return err
}

func (s *Service) GetByUser(uid uuid.UUID) (dms []domain.Reminder, err error) {
	dms, err = s.repo.GetByUser(uid)

	return dms, err
}

func (s *Service) GetByTask(uid uuid.UUID) (dms []domain.Reminder, err error) {
	dms, err = s.repo.GetByTask(uid)

	return dms, err
}

func (s *Service) Get(uid uuid.UUID) (dm domain.Reminder, err error) {
	return s.repo.Get(uid)
}

func (s *Service) GetRemindersNames(ctx context.Context, uids []uuid.UUID) (withName []domain.Reminder, err error) {
	if len(uids) == 0 {
		return withName, nil
	}

	return s.repo.GetRemindersNames(ctx, uids)
}

func (s *Service) GetPeople(r domain.Reminder) (emails []string, err error) {
	user, fu := s.dict.FindUserByUUID(r.CreatedByUUID)
	if !fu {
		return emails, fmt.Errorf("reminder createt user not found by uuid: %s", r.CreatedByUUID.String())
	}

	emails = append(emails, user.Email)

	if r.UserUUID != nil {
		cuser, cu := s.dict.FindUserByUUID(*r.UserUUID)
		if !cu {
			return emails, fmt.Errorf("reminder user not found by uuid: %s", r.UserUUID.String())
		}

		emails = append(emails, cuser.Email)
	}

	return emails, nil
}
