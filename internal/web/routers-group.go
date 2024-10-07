package web

import (
	"context"
	"errors"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

/// GROUPS

func (a *Web) PostCompanyUUIDGroup(ctx context.Context, request oapi.PostCompanyUUIDGroupRequestObject) (oapi.PostCompanyUUIDGroupResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	company, found := a.app.DictionaryService.FindCompany(request.UUID)
	if !found {
		return nil, errors.New("компания не найдена")
	}

	group := domain.NewGroup(request.Body.Name, company.FederationUUID, company.UUID)

	err := a.app.FederationService.CreateGroup(group)
	if err != nil {
		return nil, err
	}

	return oapi.PostCompanyUUIDGroup200JSONResponse{
		Uuid: group.UUID,
	}, nil
}

func (a *Web) GetCompanyUUIDGroup(ctx context.Context, request oapi.GetCompanyUUIDGroupRequestObject) (oapi.GetCompanyUUIDGroupResponseObject, error) {
	dmns, err := a.app.FederationService.GetCompanyGroups(ctx, request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetCompanyUUIDGroup200JSONResponse{
		Count: len(dmns),
		Items: lo.Map(dmns, func(item domain.Group, index int) dto.GroupDTO {
			return dto.GroupDTO{
				UUID: item.UUID,
				Name: item.Name,

				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,

				UserUUIDS: item.UsersUUIDS,
			}
		}),
	}, nil
}

func (a *Web) PatchCompanyUUIDGroupEntityUUID(ctx context.Context, request oapi.PatchCompanyUUIDGroupEntityUUIDRequestObject) (oapi.PatchCompanyUUIDGroupEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.ChangeGroupName(request.EntityUUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return oapi.PatchCompanyUUIDGroupEntityUUID200Response{}, nil
}

func (a *Web) DeleteCompanyUUIDGroupEntityUUID(ctx context.Context, request oapi.DeleteCompanyUUIDGroupEntityUUIDRequestObject) (oapi.DeleteCompanyUUIDGroupEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteGroup(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCompanyUUIDGroupEntityUUID200Response{}, nil
}

func (a *Web) GetGroupUUIDUser(ctx context.Context, request oapi.GetGroupUUIDUserRequestObject) (oapi.GetGroupUUIDUserResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmns, err := a.app.FederationService.GetGroupUsers(ctx, request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetGroupUUIDUser200JSONResponse{
		Count: len(dmns),
		Items: lo.Map(dmns, func(item domain.User, index int) dto.UserDTO {
			return dto.UserDTO{
				UUID:  item.UUID,
				Email: item.Email,
				Name:  item.Name,
				Lname: item.Lname,
				Pname: item.Pname,
				Phone: item.Phone,
			}
		}),
	}, nil
}

func (a *Web) PostGroupUUIDUser(ctx context.Context, request oapi.PostGroupUUIDUserRequestObject) (oapi.PostGroupUUIDUserResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.AddUserToGroup(request.Body.Uuid, request.UUID, claims.Email, claims.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.PostGroupUUIDUser200Response{}, nil
}

func (a *Web) DeleteGroupUUIDUser(ctx context.Context, request oapi.DeleteGroupUUIDUserRequestObject) (oapi.DeleteGroupUUIDUserResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.RemoveUserFromGroups(request.UUID, request.Body.Uuid)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteGroupUUIDUser200Response{}, nil
}
