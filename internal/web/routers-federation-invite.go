package web

import (
	"context"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

func (a *Web) GetFederationUUIDInvite(ctx context.Context, request oapi.GetFederationUUIDInviteRequestObject) (oapi.GetFederationUUIDInviteResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dms, err := a.app.FederationService.GetInvites(request.UUID)
	if err != nil {
		return nil, err
	}

	dtos := lo.Map(dms, func(item domain.Invite, index int) dto.InviteDTO {
		var company *dto.CompanyDTOs
		var federation *dto.FederationDTOs
		if item.CompanyUUID != nil {
			companyDto, f := a.app.DictionaryService.FindCompany(*item.CompanyUUID)
			if f {
				company = &dto.CompanyDTOs{
					UUID: companyDto.UUID,
					Name: companyDto.Name,
				}
			}
		}

		federationDto, f := a.app.DictionaryService.FindFederation(item.FederationUUID)
		if f {
			federation = &dto.FederationDTOs{
				UUID: federationDto.UUID,
				Name: federationDto.Name,
			}
		}

		return dto.InviteDTO{
			UUID:           item.UUID,
			Email:          item.Email,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			Federation:     federation,
			Company:        company,
			AcceptedAt:     item.AcceptedAt,
			DeclinedAt:     item.DeclinedAt,

			CreatedAt: item.CreatedAt,
		}
	})

	return oapi.GetFederationUUIDInvite200JSONResponse{
		Count: len(dms),
		Items: dtos,
	}, nil
}

func (a *Web) PostFederationUUIDInvite(ctx context.Context, request oapi.PostFederationUUIDInviteRequestObject) (oapi.PostFederationUUIDInviteResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	invite := domain.NewInvite(request.Body.Email, request.UUID, request.Body.CompanyUuid)

	err := a.app.FederationService.InviteUser(invite)
	if err != nil {
		return nil, err
	}

	return oapi.PostFederationUUIDInvite200JSONResponse{
		Uuid: invite.UUID,
	}, nil
}

func (a *Web) DeleteFederationUUIDInviteEntityUUID(ctx context.Context, request oapi.DeleteFederationUUIDInviteEntityUUIDRequestObject) (oapi.DeleteFederationUUIDInviteEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteInvite(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteFederationUUIDInviteEntityUUID200Response{}, nil
}
