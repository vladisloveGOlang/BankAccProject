package web

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func initOpenAPIFederationRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oFederation").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) PostFederation(ctx context.Context, request oapi.PostFederationRequestObject) (oapi.PostFederationResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.GateService.FederationCreate(claims.UUID)
	if err != nil {
		return nil, err
	}

	federation := domain.NewFederation(request.Body.Name, claims.Email, claims.UUID)

	err = a.app.FederationService.CreateFederation(federation)
	if err != nil {
		return nil, err
	}

	return oapi.PostFederation200JSONResponse{
		Uuid: federation.UUID,
	}, nil
}

func (a *Web) GetFederationUUID(ctx context.Context, request oapi.GetFederationUUIDRequestObject) (oapi.GetFederationUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	federation, err := a.app.FederationService.GetFederation(request.UUID)
	if err != nil {
		return nil, err
	}

	userGroups, err := a.app.FederationService.GetUsersGroups(ctx, lo.Map(federation.Users, func(item domain.FederationUser, index int) uuid.UUID {
		return item.UUID
	}))
	if err != nil {
		return nil, err
	}

	fderationUserDto := lo.Map(federation.Users, func(item domain.FederationUser, index int) dto.FederationUserDto {
		return *dto.NewFederationUserDto(item, userGroups, a.app.ProfileService)
	})

	companiesDTO := lo.Map(federation.Companies, func(item domain.Company, index int) dto.CompanyDTO {
		return dto.CompanyDTO{
			UUID:           item.UUID,
			Name:           item.Name,
			FederationUUID: item.FederationUUID,
			CreatedAt:      item.CreatedAt,
			DeletedAt:      item.DeletedAt,
			UpdatedAt:      item.UpdatedAt,
			UsersTotal:     item.UserTotal,
			ProjectsTotal:  item.ProjectsTotal,
		}
	})

	return oapi.GetFederationUUID200JSONResponse(dto.FederationDTO{
		UUID:      federation.UUID,
		Name:      federation.Name,
		CreatedAt: federation.CreatedAt,
		DeletedAt: federation.DeletedAt,

		Users:      fderationUserDto,
		UsersTotal: federation.UsersTotal,

		CompaniesTotal: len(companiesDTO),
		Companies:      companiesDTO,
	}), nil
}

func (a *Web) DeleteFederationUUID(ctx context.Context, request oapi.DeleteFederationUUIDRequestObject) (oapi.DeleteFederationUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteFederation(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteFederationUUID200Response{}, nil
}

func (a *Web) PostFederationUUIDUser(ctx context.Context, request oapi.PostFederationUUIDUserRequestObject) (oapi.PostFederationUUIDUserResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	fu := domain.NewFederationUser(request.UUID, request.Body.UserUuid)

	err := a.app.FederationService.AddUser(*fu)
	if err != nil {
		return nil, err
	}

	return oapi.PostFederationUUIDUser200JSONResponse{
		Uuid: fu.UUID,
	}, nil
}

func (a *Web) DeleteFederationUUIDUserUserUUID(ctx context.Context, request oapi.DeleteFederationUUIDUserUserUUIDRequestObject) (oapi.DeleteFederationUUIDUserUserUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteUser(request.UUID, request.UserUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteFederationUUIDUserUserUUID200Response{}, nil
}

// PatchFederationUUIDName implements ofederation.StrictServerInterface.
func (a *Web) PatchFederationUUIDName(ctx context.Context, request oapi.PatchFederationUUIDNameRequestObject) (oapi.PatchFederationUUIDNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.ChangeName(request.UUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return oapi.PatchFederationUUIDName200Response{}, nil
}
