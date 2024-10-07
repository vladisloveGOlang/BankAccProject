package gates

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (a *Service) FederationCreate(userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	if len(fUUIDs) == 0 {
		return nil
	}

	federationCounts := len(fUUIDs)

	if federationCounts >= a.federationLimit {
		logrus.Debug(federationCounts, a.federationLimit)
		return fmt.Errorf("превышен лимит федераций (максимум %v)", a.federationLimit)
	}

	return nil
}

func (a *Service) FederationDelete(federationUUID, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена")
	}

	federation, found := a.dict.FindFederation(federationUUID)
	if !found {
		return fmt.Errorf("федерация не найдена")
	}

	if federation.CreatedByUUID == nil || *federation.CreatedByUUID == userUUID {
		return fmt.Errorf("федерцию может удалить только создатель")
	}

	return nil
}

func (a *Service) FederationPatch(federationUUID, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена")
	}

	federation, found := a.dict.FindFederation(federationUUID)
	if !found {
		return fmt.Errorf("федерация не найдена")
	}

	if federation.CreatedByUUID == nil || *federation.CreatedByUUID == userUUID {
		return fmt.Errorf("федерцию может изменить только создатель")
	}

	return nil
}

func (a *Service) FederationAddUser(federationUUID, _, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена")
	}

	return nil
}

func (a *Service) FederationRemoveUser(federationUUID, _, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена")
	}

	return nil
}
