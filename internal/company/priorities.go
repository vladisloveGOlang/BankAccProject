package company

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

func (s *Service) CreateCompanyPriority(dm domain.CompanyPriority) (err error) {
	err = s.repo.CreateCompanyPriority(dm)

	return err
}

func (s *Service) GetCompanyPriority(uid uuid.UUID) (dm domain.CompanyPriority, err error) {
	orm, err := s.repo.GetCompanyPriority(uid)
	if err != nil {
		return dm, err
	}

	return domain.CompanyPriority{
		UUID:        orm.UUID,
		Name:        orm.Name,
		Number:      orm.Number,
		Color:       orm.Color,
		CompanyUUID: orm.CompanyUUID,
	}, err
}

func (s *Service) GetCompanyPriorities(companyUUID uuid.UUID) (tags []domain.CompanyPriority, err error) {
	orms, err := s.repo.GetCompanyPriorities(companyUUID)
	if err != nil {
		return tags, err
	}

	return lo.Map(orms, func(orm CompanyPriority, _ int) domain.CompanyPriority {
		return domain.CompanyPriority{
			UUID:        orm.UUID,
			Name:        orm.Name,
			Number:      orm.Number,
			Color:       orm.Color,
			CompanyUUID: orm.CompanyUUID,
		}
	}), err
}

func (s *Service) UpdateCompanyPriority(uid uuid.UUID, name, color string) (err error) {
	err = s.repo.UpdateCompanyPriority(uid, name, color)
	return err
}

func (s *Service) DeleteCompanyPriority(uid uuid.UUID) (err error) {
	err = s.repo.DeleteCompanyPriority(uid)
	return err
}
