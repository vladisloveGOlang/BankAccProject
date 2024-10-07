package web

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

func (a *Web) PostCompany(ctx context.Context, request oapi.PostCompanyRequestObject) (oapi.PostCompanyResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	company := domain.NewCompany(request.Body.Name, request.Body.FederationUuid, claims.Email, claims.UUID)

	err := a.app.FederationService.CreateCompany(company, true)
	if err != nil {
		return nil, err
	}

	return oapi.PostCompany200JSONResponse{
		Uuid: company.UUID,
	}, nil
}

func (a *Web) GetCompanyUUID(ctx context.Context, request oapi.GetCompanyUUIDRequestObject) (oapi.GetCompanyUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	item, err := a.app.FederationService.GetCompany(request.UUID)
	if err != nil {
		return nil, err
	}

	users := lo.Map(item.Users, func(item domain.CompanyUser, index int) dto.UserDTO {
		return dto.UserDTO{
			UUID:  item.User.UUID,
			Name:  item.User.Name,
			Lname: item.User.Lname,
			Pname: item.User.Pname,
			Email: item.User.Email,
			Phone: item.User.Phone,
		}
	})

	projects := lo.Map(item.Projects, func(item domain.Project, index int) dto.ProjectDTOs {
		return dto.ProjectDTOs{
			UUID: item.UUID,
			Name: item.Name,
		}
	})

	// @todo: refactor
	companyPriorities, err := a.app.CompanyService.GetCompanyPriorities(item.UUID)
	if err != nil {
		return nil, err
	}
	companyPrioritiesDto := lo.Map(companyPriorities, func(item domain.CompanyPriority, index int) dto.CompanyPriorityDTO {
		return dto.CompanyPriorityDTO{
			UUID:   item.UUID,
			Name:   item.Name,
			Number: item.Number,
			Color:  item.Color,
		}
	})

	return oapi.GetCompanyUUID200JSONResponse(dto.CompanyDTO{
		UUID:           item.UUID,
		Name:           item.Name,
		FederationUUID: item.FederationUUID,

		UsersTotal: len(users),
		Users:      users,

		ProjectsTotal: len(projects),
		Projects:      projects,
		Priorities:    companyPrioritiesDto,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}), nil
}

func (a *Web) PatchCompanyUUIDName(ctx context.Context, request oapi.PatchCompanyUUIDNameRequestObject) (oapi.PatchCompanyUUIDNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.ChangeCompanyName(request.UUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return oapi.PatchCompanyUUIDName200Response{}, nil
}

func (a *Web) DeleteCompanyUUID(ctx context.Context, request oapi.DeleteCompanyUUIDRequestObject) (oapi.DeleteCompanyUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteCompany(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCompanyUUID200Response{}, nil
}

func (a *Web) PostCompanyUUIDUser(ctx context.Context, request oapi.PostCompanyUUIDUserRequestObject) (oapi.PostCompanyUUIDUserResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	federationDTO, err := a.app.FederationService.GetCompanyFederation(ctx, request.UUID)
	if err != nil {
		return nil, err
	}

	if federationDTO.UUID == uuid.Nil {
		return nil, errors.New("федерация не найдена")
	}

	fu := domain.NewCompanyUser(federationDTO.UUID, request.UUID, request.Body.UserUuid)

	err = a.app.FederationService.AddUserToCompany(*fu)
	if err != nil {
		return nil, err
	}

	return oapi.PostCompanyUUIDUser200JSONResponse{
		Uuid: fu.UUID,
	}, nil
}

func (a *Web) DeleteCompanyUUIDUserUserUUID(ctx context.Context, request oapi.DeleteCompanyUUIDUserUserUUIDRequestObject) (oapi.DeleteCompanyUUIDUserUserUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteUserFromCompany(request.UUID, request.UserUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCompanyUUIDUserUserUUID200Response{}, nil
}
