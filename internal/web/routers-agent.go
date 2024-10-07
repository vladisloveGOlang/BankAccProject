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

func (a *Web) DeleteFederationUUIDAgentEntityUUID(ctx context.Context, request oapi.DeleteFederationUUIDAgentEntityUUIDRequestObject) (oapi.DeleteFederationUUIDAgentEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.AgentsService.Delete(ctx, request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteFederationUUIDAgentEntityUUID200Response{}, nil
}

func (a *Web) GetFederationUUIDAgent(ctx context.Context, request oapi.GetFederationUUIDAgentRequestObject) (oapi.GetFederationUUIDAgentResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	filter := domain.AgentFilter{
		Offset:         request.Params.Offset,
		Limit:          request.Params.Limit,
		FederationUUID: request.UUID,
	}

	dms, total, err := a.app.AgentsService.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return oapi.GetFederationUUIDAgent200JSONResponse{
		Count: len(dms),
		Items: lo.Map(dms, func(item domain.Agent, i int) dto.AgentDTO {
			return dto.AgentDTO{
				UUID:           item.UUID,
				FederationUUID: item.FederationUUID,
				CompanyUUID:    item.CompanyUUID,

				Name: item.Name,
				Contacts: lo.Map(item.Contacts, func(c domain.AgentContacts, _ int) dto.AgentContactsDTO {
					return dto.AgentContactsDTO{
						Type: c.Type,
						Val:  c.Val,
					}
				}),
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}
		}),
		Total: total,
	}, nil
}

func (a *Web) PatchFederationUUIDAgentEntityUUID(ctx context.Context, request oapi.PatchFederationUUIDAgentEntityUUIDRequestObject) (oapi.PatchFederationUUIDAgentEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	ac := []domain.AgentContacts{}
	for c := range request.Body.Contacts {
		if request.Body.Contacts[c].Type == "" || request.Body.Contacts[c].Value == "" {
			return nil, errors.New("тип и значение контакта не могут быть пустыми")
		}

		ac = append(ac, domain.AgentContacts{
			Type: request.Body.Contacts[c].Type,
			Val:  request.Body.Contacts[c].Value,
		})
	}

	err := a.app.AgentsService.Update(ctx, &domain.Agent{
		UUID:     request.EntityUUID,
		Name:     request.Body.Name,
		Contacts: ac,
	})
	if err != nil {
		return nil, err
	}

	return oapi.PatchFederationUUIDAgentEntityUUID200Response{}, nil
}

func (a *Web) PostFederationUUIDAgent(ctx context.Context, request oapi.PostFederationUUIDAgentRequestObject) (oapi.PostFederationUUIDAgentResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	ac := []domain.AgentContacts{}
	for c := range request.Body.Contacts {
		if request.Body.Contacts[c].Type == "" || request.Body.Contacts[c].Value == "" {
			return nil, errors.New("тип и значение контакта не могут быть пустыми")
		}

		ac = append(ac, domain.AgentContacts{
			Type: request.Body.Contacts[c].Type,
			Val:  request.Body.Contacts[c].Value,
		})
	}

	dm := domain.NewAgent(request.UUID, request.Body.CompanyUuid, domain.Me{
		Email: claims.Email,
		UUID:  claims.UUID,
	}, request.Body.Name, ac)

	err := a.app.AgentsService.Create(ctx, dm)
	if err != nil {
		return nil, err
	}

	return oapi.PostFederationUUIDAgent200JSONResponse{
		Uuid: dm.UUID,
	}, nil
}
