package web

import "github.com/krisch/crm-backend/internal/legalEntities"

type handler struct {
	service *legalEntities.LegalEntitiesService
}

func (h handler) GetLegalEntities()
