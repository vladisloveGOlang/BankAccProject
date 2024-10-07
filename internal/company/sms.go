package company

import (
	"github.com/google/uuid"
)

func (s *Service) CreateSmsOptions(uid uuid.UUID, so SmsOptions) (err error) {
	err = s.repo.UpdateSmsOptions(uid, so)

	return err
}

func (s *Service) GetSmsOptions(uid uuid.UUID) (so SmsOptions, err error) {
	orm, err := s.repo.GetSmsOptions(uid)
	if err != nil {
		return so, err
	}

	return orm.SmsOptions, err
}
