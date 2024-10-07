package gates

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

func (a *Service) CompanyCreate(federationUUID, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена")
	}

	federation, found := a.dict.FindFederation(federationUUID)

	if !found {
		return fmt.Errorf("федерация не существует")
	}

	companiesCounts := federation.CompaniesTotal

	if companiesCounts >= a.commentsLimit {
		return fmt.Errorf("превышен лимит компаний (максимум %v)", a.commentsLimit)
	}

	return nil
}

func (a *Service) CompanyDelete(companyUUID, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, companyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}

func (a *Service) CompanyPatch(companyUUID, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, companyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}

func (a *Service) CompanyAddUser(companyUUID, _, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, companyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}

func (a *Service) CompanyRemoveUser(companyUUID, _, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, companyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}
