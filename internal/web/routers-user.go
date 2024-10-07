package web

import (
	"context"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

func (a *Web) GetUser(ctx context.Context, request oapi.GetUserRequestObject) (oapi.GetUserResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.GateService.UsersSearch(request.Body.FederationUuid, claims.UUID)
	if err != nil {
		return nil, err
	}

	search := domain.SearchUser{
		Search:         request.Body.Search,
		FederationUUID: request.Body.FederationUuid,
		CompanyUUID:    request.Body.CompanyUuid,
	}

	dmns, err := a.app.FederationService.SearchUserInDictionary(search)
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	dtos := lo.Map(dmns, func(item domain.User, index int) dto.UserDTO {
		return dto.NewUserDto(item, a.app.ProfileService)
	})

	return oapi.GetUser200JSONResponse{
		Count: len(dtos),
		Items: dtos,
	}, nil
}
