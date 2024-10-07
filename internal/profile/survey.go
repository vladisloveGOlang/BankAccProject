package profile

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *Service) CreateSurvey(survey domain.Survey) error {
	return s.repo.CreateSurvey(survey)
}

func (s *Service) DeleteSurvey(uid uuid.UUID) error {
	return s.repo.DeleteSurvey(uid)
}

func (s *Service) GetSurvey(uid uuid.UUID) (domain.Survey, error) {
	orm, err := s.repo.GetSurvey(uid)
	if err != nil {
		return domain.Survey{}, err
	}

	// @todo: refactor
	user, f := s.dict.FindUserByUUID(orm.UserUUID)
	dmUser := domain.User{}
	if !f {
		logrus.WithError(err).Error("failed to find user")
	} else {
		dmUser = domain.User{
			Name:     user.Name,
			Lname:    user.Lname,
			Pname:    user.Pname,
			UUID:     user.UUID,
			Email:    user.Email,
			Phone:    user.Phone,
			HasPhoto: user.HasPhoto,
		}

		var photo *domain.ProfilePhotoDTO
		if user.Photo != nil {
			photo = &domain.ProfilePhotoDTO{
				Small:  user.Photo.Small,
				Medium: user.Photo.Medium,
				Large:  user.Photo.Large,
			}
		}
		dmUser.Photo = photo
	}

	return domain.Survey{
		UUID:      orm.UUID,
		Name:      orm.Name,
		User:      dmUser,
		Body:      orm.Body,
		CreatedAt: orm.CreatedAt,
		UpdatedAt: orm.UpdatedAt,
		DeletedAt: orm.DeletedAt,
	}, nil
}

func (s *Service) GetSurveyByUserUUID(uid uuid.UUID) ([]domain.Survey, error) {
	orms, err := s.repo.GetSurveyByUserUUID(uid)
	if err != nil {
		return []domain.Survey{}, err
	}

	return lo.Map(orms, func(orm Survey, _ int) domain.Survey {
		return domain.Survey{
			UUID:      orm.UUID,
			Name:      orm.Name,
			Body:      orm.Body,
			CreatedAt: orm.CreatedAt,
			UpdatedAt: orm.UpdatedAt,
		}
	}), nil
}
