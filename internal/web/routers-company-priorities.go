package web

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

func (a *Web) DeleteCompanyUUIDPrioritiesEntityUUID(ctx context.Context, request oapi.DeleteCompanyUUIDPrioritiesEntityUUIDRequestObject) (oapi.DeleteCompanyUUIDPrioritiesEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CompanyService.DeleteCompanyPriority(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCompanyUUIDPrioritiesEntityUUID200Response{}, nil
}

func (a *Web) GetCompanyUUIDPriorities(ctx context.Context, request oapi.GetCompanyUUIDPrioritiesRequestObject) (oapi.GetCompanyUUIDPrioritiesResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	cps, err := a.app.CompanyService.GetCompanyPriorities(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetCompanyUUIDPriorities200JSONResponse{
		Count: len(cps),
		Items: lo.Map(cps, func(cp domain.CompanyPriority, _ int) dto.CompanyPriorityDTO {
			return dto.CompanyPriorityDTO{
				UUID:   cp.UUID,
				Name:   cp.Name,
				Color:  cp.Color,
				Number: cp.Number,
			}
		}),
	}, nil
}

func (a *Web) PatchCompanyUUIDPrioritiesEntityUUID(ctx context.Context, request oapi.PatchCompanyUUIDPrioritiesEntityUUIDRequestObject) (oapi.PatchCompanyUUIDPrioritiesEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CompanyService.UpdateCompanyPriority(request.EntityUUID, request.Body.Name, request.Body.Color)
	if err != nil {
		return nil, err
	}

	return oapi.PatchCompanyUUIDPrioritiesEntityUUID200Response{}, nil
}

func (a *Web) PostCompanyUUIDPriorities(ctx context.Context, request oapi.PostCompanyUUIDPrioritiesRequestObject) (oapi.PostCompanyUUIDPrioritiesResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dm := domain.CompanyPriority{
		UUID:        uuid.New(),
		CompanyUUID: request.UUID,
		Name:        request.Body.Name,
		Number:      request.Body.Number,
		Color:       request.Body.Color,
	}

	err := a.app.CompanyService.CreateCompanyPriority(dm)
	if err != nil {
		return nil, err
	}

	return oapi.PostCompanyUUIDPriorities200JSONResponse{
		Uuid: dm.UUID,
	}, nil
}
