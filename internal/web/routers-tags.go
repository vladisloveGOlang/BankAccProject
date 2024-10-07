package web

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

func (a *Web) PostTag(ctx context.Context, request oapi.PostTagRequestObject) (oapi.PostTagResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	company, f := a.app.DictionaryService.FindCompany(request.Body.CompanyUuid)
	if !f {
		return nil, fmt.Errorf("company not found. uuid: %s", request.Body.CompanyUuid)
	}

	tag := domain.Tag{
		UUID:  uuid.New(),
		Name:  request.Body.Name,
		Color: request.Body.Color,

		FederationUUID: company.FederationUUID,
		CompanyUUID:    company.UUID,

		CreatedBy: domain.User{
			UUID:  claims.UUID,
			Email: claims.Email,
		},
	}

	err := a.app.CompanyService.CreateTag(tag)
	if err != nil {
		return nil, err
	}

	return oapi.PostTag200JSONResponse{
		Uuid: tag.UUID,
	}, nil
}

func (a *Web) PatchTagUUID(ctx context.Context, request oapi.PatchTagUUIDRequestObject) (oapi.PatchTagUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CompanyService.UpdateTag(request.UUID, request.Body.Name, request.Body.Color)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTagUUID200Response{}, nil
}

func (a *Web) GetTag(ctx context.Context, request oapi.GetTagRequestObject) (oapi.GetTagResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	tags, err := a.app.CompanyService.GetTags(request.Params.CompanyUuid)
	if err != nil {
		return nil, err
	}

	return oapi.GetTag200JSONResponse{
		Count: len(tags),
		Items: lo.Map(tags, func(tag domain.Tag, _ int) dto.TagDTO {
			return dto.TagDTO{
				UUID:  tag.UUID,
				Name:  tag.Name,
				Color: tag.Color,
				CreatedBy: dto.UserDTO{
					UUID:  tag.CreatedBy.UUID,
					Name:  tag.CreatedBy.Name,
					Lname: tag.CreatedBy.Lname,
					Pname: tag.CreatedBy.Pname,
					Email: tag.CreatedBy.Email,
				},
			}
		}),
	}, nil
}

func (a *Web) DeleteTagUUID(ctx context.Context, request oapi.DeleteTagUUIDRequestObject) (oapi.DeleteTagUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CompanyService.DeleteTag(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteTagUUID200Response{}, nil
}
