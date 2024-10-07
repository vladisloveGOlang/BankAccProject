package federation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
)

func (s *Service) CreateCompany(company *domain.Company, createCatalogs bool) (err error) {
	errs, ok := helpers.ValidationStruct(company)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return err
	}

	err = s.repo.CreateCompany(company)
	if err != nil {
		return err
	}

	cu := domain.NewCompanyUser(company.FederationUUID, company.UUID, company.CreatedByUUID)

	err = s.AddUserToCompany(*cu)
	if err != nil {
		return err
	}

	if createCatalogs {
		uid, err := s.catalogs.CreateCompanyCatalog(company.FederationUUID, company.UUID, company.CreatedByUUID, company.CreatedBy)
		if err != nil {
			return err
		}

		_, err = s.catalogs.CreateUsersCatalog(uid, company.FederationUUID, company.UUID, company.CreatedByUUID, company.CreatedBy)
		if err != nil {
			return err
		}

		_, err = s.catalogs.CreateGoodsCatalog(company.FederationUUID, company.UUID, company.CreatedByUUID, company.CreatedBy)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) GetCompany(uid uuid.UUID) (dm domain.Company, err error) {
	dm, err = s.repo.GetCompany(uid)
	if err != nil {
		return dm, err
	}

	users, err := s.repo.GetCompanyUsers(uid)
	if err != nil {
		return dm, err
	}

	dm.Users = users

	//

	projects, err := s.repo.GetProjectsByCompany(uid)
	if err != nil {
		return dm, err
	}

	dm.Projects = projects

	return dm, err
}

func (s *Service) ChangeCompanyName(uid uuid.UUID, name string) (err error) {
	c := domain.NewCompanyByUUID(uid)

	err = c.ChangeName(name)
	if err != nil {
		return err
	}

	err = s.repo.ChangeCompanyField(c.UUID, "name", c.Name)

	return err
}

func (s *Service) DeleteCompany(uid uuid.UUID) (err error) {
	err = s.repo.DeleteCompany(uid)
	if err != nil {
		return err
	}

	err = s.repo.DeleteProjects(&uid, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCompaniesByUser(ctx context.Context, userUUID uuid.UUID) (items []domain.Company, err error) {
	defer Span(NewSpan(ctx, "GetCompaniesByUser"))()

	items, err = s.repo.GetCompaniesByUser(userUUID)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) AddUserToCompany(fu domain.CompanyUser) (err error) {
	err = s.repo.AddUserToCompany(fu)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteUserFromCompany(companyUUID, userUUID uuid.UUID) (err error) {
	err = s.repo.DeleteUserFromCompany(companyUUID, userUUID)
	if err != nil {
		return err
	}

	return err
}
