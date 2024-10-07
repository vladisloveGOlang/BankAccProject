package aggregates

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *Service) GetProject(_ context.Context, uid uuid.UUID) (dto.ProjectDTO, error) {
	dmn, err := s.federationService.GetProject(uid)
	if err != nil {
		return dto.ProjectDTO{}, err
	}

	graph := make(map[string][]string)
	if dmn.StatusGraph != nil {
		graph = dmn.StatusGraph.Graph
	}

	federation, found := s.dictionaryService.FindFederation(dmn.FederationUUID)
	if !found {
		return dto.ProjectDTO{}, dto.NotFoundErr("федерация не найдена")
	}

	company, found := s.dictionaryService.FindCompany(dmn.CompanyUUID)
	if !found {
		return dto.ProjectDTO{}, dto.NotFoundErr("компания не найдена")
	}

	allowSort := s.ts.GetSortFields(dmn.UUID)

	responsibleBy, _ := s.dictionaryService.FindUser(dmn.ResponsibleBy)

	//

	statuses, err := s.federationService.GetProjectStatuses(dmn.UUID)
	if err != nil {
		return dto.ProjectDTO{}, err
	}
	StatusesDTO := lo.Map(statuses, func(cp domain.ProjectStatus, _ int) dto.ProjectStatusDTO {
		return dto.ProjectStatusDTO{
			UUID:        cp.UUID,
			Name:        cp.Name,
			Color:       cp.Color,
			Number:      cp.Number,
			Description: cp.Description,
			Edit:        cp.Edit,
		}
	})

	status, found := lo.Find(StatusesDTO, func(item dto.ProjectStatusDTO) bool {
		return item.Number == dmn.StatusCode
	})
	if !found {
		logrus.
			WithField("project_uuid", dmn.UUID).
			WithField("number", dmn.StatusCode).
			Error("status not found by number")
	}

	//
	options := dto.ProjectOptionsDTO(dmn.Options)

	dt := dto.ProjectDTO{
		UUID:        dmn.UUID,
		Name:        dmn.Name,
		Description: dmn.Description,

		FieldsTotal: len(dmn.Fields),
		Fields: helpers.Map(dmn.Fields, func(item domain.CompanyField, index int) dto.ProjectFieldDTO {
			return dto.ProjectFieldDTO{
				UUID:               item.UUID,
				Name:               item.Name,
				Description:        item.Description,
				Hash:               item.Hash,
				DataType:           int(item.DataType),
				RequiredOnStatuses: item.RequiredOnStatuses,
				Style:              item.Style,
				DataDesc:           item.FieldTypeDesc(),
			}
		}),

		FederationUUID: dmn.FederationUUID,
		Federation: dto.FederationDTOs{
			UUID: federation.UUID,
			Name: federation.Name,
		},
		CompanyUUID: dmn.CompanyUUID,
		Company: dto.CompanyDTOs{
			UUID: company.UUID,
			Name: company.Name,
		},

		StatusGraph: &graph,
		Options:     &options,

		Users: helpers.Map(dmn.Users, func(item domain.ProjectUser, index int) dto.ProjectUserDto {
			return dto.ProjectUserDto{
				UUID: item.UUID,
				User: dto.UserDTO{
					UUID:     item.User.UUID,
					Name:     item.User.Name,
					Lname:    item.User.Lname,
					Pname:    item.User.Pname,
					Phone:    item.User.Phone,
					HasPhoto: item.User.Photo != nil,
					Email:    item.User.Email,
				},
			}
		}),

		AllowSort: allowSort,

		ResponsibleBy: responsibleBy,

		CreatedAt: dmn.CreatedAt,
		UpdatedAt: dmn.UpdatedAt,

		Status:     status,
		StatusCode: dmn.StatusCode,
		Statuses:   &StatusesDTO,
	}

	return dt, nil
}
