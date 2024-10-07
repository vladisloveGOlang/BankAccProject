package web

import (
	"context"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
)

func (a *Web) PostCompanyUUIDFields(ctx context.Context, request oapi.PostCompanyUUIDFieldsRequestObject) (oapi.PostCompanyUUIDFieldsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	pf := &domain.CompanyField{
		CompanyUUID: request.UUID,
		Name:        request.Body.Name,
		Description: request.Body.Description,
		DataType:    request.Body.DataType,
		Icon:        request.Body.Icon,
	}

	dt, err := a.app.FederationService.CreateCompanyField(pf)
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	return oapi.PostCompanyUUIDFields200JSONResponse{
		Uuid:            dt.UUID.String(),
		Hash:            dt.Hash,
		Type:            domain.FieldDataType(dt.DataType),
		TypeDescription: pf.FieldTypeDesc(),
		Icon:            dt.Icon,
	}, nil
}

func (a *Web) PutCompanyUUIDFieldsEntityUUID(ctx context.Context, request oapi.PutCompanyUUIDFieldsEntityUUIDRequestObject) (oapi.PutCompanyUUIDFieldsEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	pf := &domain.CompanyField{
		CompanyUUID:        request.UUID,
		UUID:               request.EntityUUID,
		Name:               request.Body.Name,
		Description:        request.Body.Description,
		Icon:               request.Body.Icon,
		RequiredOnStatuses: request.Body.RequiredOnStatuses,
	}

	err := a.app.FederationService.PutCompanyField(pf)
	if err != nil {
		return nil, err
	}

	return oapi.PutCompanyUUIDFieldsEntityUUID200Response{}, nil
}

func (a *Web) GetCompanyUUIDFields(ctx context.Context, request oapi.GetCompanyUUIDFieldsRequestObject) (oapi.GetCompanyUUIDFieldsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	items, err := a.app.FederationService.GetCompanyFields(request.UUID)
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	fields := helpers.Map(items, func(item domain.CompanyField, index int) dto.CompanyFieldDTO {
		return dto.CompanyFieldDTO{
			UUID:         item.UUID,
			Name:         item.Name,
			Description:  item.Description,
			Hash:         item.Hash,
			Icon:         item.Icon,
			DataType:     int(item.DataType),
			DataDesc:     item.FieldTypeDesc(),
			ProjectsUUID: item.ProjectUUID,

			TasksTotal:        item.TasksTotal,
			TasksFilled:       item.TasksFilled,
			TasksActiveFilled: item.TasksActiveFilled,
			TasksActiveTotal:  item.TasksActiveTotal,
		}
	})

	return oapi.GetCompanyUUIDFields200JSONResponse{
		Count: len(items),
		Items: fields,
	}, nil
}

// DeleteCompanyUUIDFieldsEntityUUID implements ofederation.StrictServerInterface.
func (a *Web) DeleteCompanyUUIDFieldsEntityUUID(ctx context.Context, request oapi.DeleteCompanyUUIDFieldsEntityUUIDRequestObject) (oapi.DeleteCompanyUUIDFieldsEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteCompanyField(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCompanyUUIDFieldsEntityUUID200Response{}, nil
}
