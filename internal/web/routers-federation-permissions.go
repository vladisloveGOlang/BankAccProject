package web

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
)

func (a *Web) PostPermissions(ctx context.Context, request oapi.PostPermissionsRequestObject) (oapi.PostPermissionsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	_, found := a.app.DictionaryService.FindFederation(request.Body.FederationUuid)
	if !found {
		return nil, dto.NotFoundErr("federation not found")
	}

	perm := domain.Permission{
		UUID:           uuid.New(),
		FederationUUID: request.Body.FederationUuid,
		UserUUID:       request.Body.UserUuid,

		Rules: domain.PermissionRules{
			FederationPatch:      request.Body.Rules.FederationPatch,
			FederationInviteUser: request.Body.Rules.FederationInviteUser,
			FederationDeleteUser: request.Body.Rules.FederationDeleteUser,

			CompanyCreate:     request.Body.Rules.CompanyCreate,
			CompanyDelete:     request.Body.Rules.CompanyDelete,
			CompanyPatch:      request.Body.Rules.CompanyPatch,
			CompanyAddUser:    request.Body.Rules.CompanyAddUser,
			CompanyDeleteUser: request.Body.Rules.CompanyDeleteUser,

			ProjectCreate:     request.Body.Rules.ProjectCreate,
			ProjectDelete:     request.Body.Rules.ProjectDelete,
			ProjectPatch:      request.Body.Rules.ProjectPatch,
			ProjectAddUser:    request.Body.Rules.ProjectAddUser,
			ProjectDeleteUser: request.Body.Rules.ProjectDeleteUser,

			TaskCreate: request.Body.Rules.TaskCreate,
			TaskDelete: request.Body.Rules.TaskDelete,
			TaskPatch:  request.Body.Rules.TaskPatch,
		},
	}

	err := a.app.GateService.CreateOrUpdatePermisson(&perm)
	if err != nil {
		return nil, err
	}

	return oapi.PostPermissions200JSONResponse{
		UpdatedAt: perm.UpdatedAt,
		UserUuid:  perm.UserUUID,
	}, nil
}

func (a *Web) GetPermissionsUUID(ctx context.Context, request oapi.GetPermissionsUUIDRequestObject) (oapi.GetPermissionsUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dm, err := a.app.GateService.GetPermisson(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetPermissionsUUID200JSONResponse{
		UserUuid:       dm.UserUUID,
		FederationUuid: dm.FederationUUID,
		Rules:          dto.PermissionRulesDTO(dm.Rules),
		CreatedAt:      dm.CreatedAt,
		UpdatedAt:      dm.UpdatedAt,
	}, nil
}

func (a *Web) DeletePermissionsUUID(ctx context.Context, request oapi.DeletePermissionsUUIDRequestObject) (oapi.DeletePermissionsUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.GateService.DeletePermission(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeletePermissionsUUID200Response{}, nil
}
