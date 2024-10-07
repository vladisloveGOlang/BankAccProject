package company

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/samber/lo"
)

type Service struct {
	repo *Repository
	dict *dictionary.Service
}

func New(repo *Repository, dict *dictionary.Service) *Service {
	return &Service{
		repo: repo,
		dict: dict,
	}
}

func (s *Service) CreateTag(tag domain.Tag) (err error) {
	err = s.repo.CreateTag(tag)
	return err
}

func (s *Service) GetTag(uid uuid.UUID) (dm domain.Tag, err error) {
	orm, err := s.repo.GetTag(uid)
	if err != nil {
		return dm, err
	}

	user, f := s.dict.FindUserByUUID(orm.CreatedByUUID)
	if !f {
		return dm, fmt.Errorf("user not found. uid: %s", orm.CreatedByUUID)
	}

	return domain.Tag{
		UUID:        orm.UUID,
		Name:        orm.Name,
		Color:       orm.Color,
		CompanyUUID: orm.CompanyUUID,
		CreatedBy: domain.User{
			Name:     user.Name,
			Lname:    user.Lname,
			Pname:    user.Pname,
			UUID:     user.UUID,
			Email:    user.Email,
			Phone:    user.Phone,
			HasPhoto: user.HasPhoto,
			Photo: &domain.ProfilePhotoDTO{
				Small:  user.Photo.Small,
				Medium: user.Photo.Medium,
				Large:  user.Photo.Large,
			},
		},
		CreatedAt: orm.CreatedAt,
		UpdatedAt: orm.UpdatedAt,
		DeletedAt: orm.DeletedAt,
	}, err
}

func (s *Service) GetTags(companyUUID uuid.UUID) (tags []domain.Tag, err error) {
	orms, err := s.repo.GetTags(companyUUID)
	if err != nil {
		return tags, err
	}

	return lo.Map(orms, func(orm CompanyTags, _ int) domain.Tag {
		user, f := s.dict.FindUserByUUID(orm.CreatedByUUID)
		if !f {
			return domain.Tag{}
		}

		photo := &domain.ProfilePhotoDTO{}
		if user.Photo != nil {
			photo = &domain.ProfilePhotoDTO{
				Small:  user.Photo.Small,
				Medium: user.Photo.Medium,
				Large:  user.Photo.Large,
			}
		}

		return domain.Tag{
			UUID:        orm.UUID,
			Name:        orm.Name,
			Color:       orm.Color,
			CompanyUUID: orm.CompanyUUID,
			CreatedBy: domain.User{
				Name:     user.Name,
				Lname:    user.Lname,
				Pname:    user.Pname,
				UUID:     user.UUID,
				Email:    user.Email,
				Phone:    user.Phone,
				HasPhoto: user.HasPhoto,
				Photo:    photo,
			},
			CreatedAt: orm.CreatedAt,
			UpdatedAt: orm.UpdatedAt,
			DeletedAt: orm.DeletedAt,
		}
	}), err
}

func (s *Service) UpdateTag(uid uuid.UUID, name, color string) (err error) {
	err = s.repo.UpdateTag(uid, name, color)
	return err
}

func (s *Service) DeleteTag(uid uuid.UUID) (err error) {
	err = s.repo.DeleteTag(uid)
	return err
}
