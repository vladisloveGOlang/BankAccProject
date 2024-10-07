package dictionary

import (
	"github.com/krisch/crm-backend/dto"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *Service) FindUsers(emails []string) ([]dto.UserDTO, []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	users := []dto.UserDTO{}

	if len(s.usersByEmail) == 0 {
		logrus.Error("users not synced")
		return users, []string{}
	}

	emails = lo.Uniq(emails)

	for _, email := range emails {
		if email != "" {
			if user, ok := s.usersByEmail[email]; ok {
				users = append(users, user)
			} else {
				logrus.Warn("user not found: ", email)
			}
		}
	}

	return users, lo.Map(users, func(u dto.UserDTO, _ int) string {
		return u.Email
	})
}
