package gates

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

func (a *Service) TaskCreate(task domain.Task, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, task.CompanyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}

func (a *Service) TaskDelete(task domain.Task, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, task.CompanyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}

func (a *Service) TaskPatch(task domain.Task, userUUID uuid.UUID) error {
	cUUIDs := a.dict.GetUserCompanies(userUUID)

	hasCompany := lo.IndexOf(cUUIDs, task.CompanyUUID)

	if hasCompany == -1 {
		return fmt.Errorf("компания не найдена")
	}

	return nil
}
