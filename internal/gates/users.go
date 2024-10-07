package gates

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

func (a *Service) UsersSearch(federationUUID, userUUID uuid.UUID) error {
	fUUIDs := a.dict.GetUserFederatons(userUUID)

	hasFederation := lo.IndexOf(fUUIDs, federationUUID)

	if hasFederation == -1 {
		return fmt.Errorf("федерация не найдена или у вас нет доступа к ней")
	}

	return nil
}
