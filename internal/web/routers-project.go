package web

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func initOpenAPIProjectRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oProject").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

// DeleteProjectUUID is a method that needs to be added to the *Web struct to implement the oapi.StrictServerInterface interface.
func (a *Web) DeleteProjectUUID(ctx context.Context, request oapi.DeleteProjectUUIDRequestObject) (oapi.DeleteProjectUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteProject(request.UUID.String())
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProjectUUID200Response{}, nil
}

func (a *Web) PostProject(ctx context.Context, request oapi.PostProjectRequestObject) (oapi.PostProjectResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	federationDTO, err := a.app.FederationService.GetCompanyFederation(ctx, request.Body.CompanyUuid)
	if err != nil {
		return nil, err
	}

	project := &domain.Project{
		UUID:           uuid.New(),
		FederationUUID: federationDTO.UUID,
		CompanyUUID:    request.Body.CompanyUuid,
		Name:           request.Body.Name,
		Description:    request.Body.Description,
		CreatedBy:      claims.Email,
		ResponsibleBy:  request.Body.ResponsibleBy,
		Options: domain.ProjectOptions{
			RequireCancelationComment: helpers.Ptr(false),
			RequireDoneComment:        helpers.Ptr(false),
			StatusEnable:              helpers.Ptr(false),
			Color:                     helpers.Ptr("#000000"),
		},
		StatusSort: request.Body.StatusSort,
		FieldsSort: request.Body.FieldsSort,
	}

	err = a.app.FederationService.CreateProgect(project)
	if err != nil {
		return nil, err
	}

	return oapi.PostProject200JSONResponse{
		Uuid: project.UUID,
	}, nil
}

func (a *Web) GetProjectUUID(ctx context.Context, request oapi.GetProjectUUIDRequestObject) (oapi.GetProjectUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmn, err := a.app.FederationService.GetProject(request.UUID)
	if err != nil {
		return nil, err
	}

	graph := make(map[string][]string)
	if dmn.StatusGraph != nil {
		graph = dmn.StatusGraph.Graph
	}

	federation, found := a.app.DictionaryService.FindFederation(dmn.FederationUUID)
	if !found {
		return nil, dto.NotFoundErr("федерация не найдена")
	}

	company, found := a.app.DictionaryService.FindCompany(dmn.CompanyUUID)
	if !found {
		return nil, dto.NotFoundErr("компания не найдена")
	}

	allowSort := a.app.TaskService.GetSortFields(dmn.UUID)

	responsibleBy, _ := a.app.DictionaryService.FindUser(dmn.ResponsibleBy)

	// Statuses
	statuses, err := a.app.FederationService.GetProjectStatuses(dmn.UUID)
	if err != nil {
		return nil, err
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

	// statistics
	statistics, fieldStatistics, err := a.app.FederationService.GetProjectStatistic(company.UUID, dmn.UUID)
	if err != nil {
		return nil, err
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

		Status:          status,
		StatusCode:      dmn.StatusCode,
		StatusUpdatedAt: dmn.StatusUpdatedAt,
		Statuses:        &StatusesDTO,

		Statistic: &dto.ProjectStatistics{
			TasksTotal:         statistics.TasksTotal,
			TasksFinishedTotal: statistics.TasksFinishedTotal,
			TaskCanceledTotal:  statistics.TasksCanceledTotal,
			TaskDeletedTotal:   statistics.TasksDeletedTotal,
			TasksActiveTotal:   statistics.TasksActiveTotal,
		},

		FieldStatistics: fieldStatistics,
	}

	return oapi.GetProjectUUID200JSONResponse(dt), nil
}

func (a *Web) GetFederationUUIDProject(ctx context.Context, request oapi.GetFederationUUIDProjectRequestObject) (oapi.GetFederationUUIDProjectResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmns, err := a.app.FederationService.GetProjects(request.UUID)
	if err != nil {
		return nil, err
	}

	items := []dto.ProjectDTO{}

	for _, dmn := range dmns {

		// @todo: to many queries here...
		// Federation
		federation, found := a.app.DictionaryService.FindFederation(dmn.FederationUUID)
		if !found {
			return nil, dto.NotFoundErr("федерация не найдена")
		}

		// Company
		company, found := a.app.DictionaryService.FindCompany(dmn.CompanyUUID)
		if !found {
			return nil, dto.NotFoundErr("компания не найдена")
		}

		// ResponsibleBy
		responsibleBy, found := a.app.DictionaryService.FindUser(dmn.ResponsibleBy)
		if !found {
			logrus.
				WithField("responsible_by", dmn.ResponsibleBy).
				WithField("project_uuid", dmn.UUID).
				Error("responsible by not found in task")
		}

		// Statuses
		statuses, err := a.app.FederationService.GetProjectStatuses(dmn.UUID)
		if err != nil {
			return nil, err
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

		items = append(items, dto.ProjectDTO{
			UUID:        dmn.UUID,
			Name:        dmn.Name,
			Description: dmn.Description,

			FieldsTotal: len(dmn.Fields),

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

			ResponsibleBy: responsibleBy,

			CreatedAt: dmn.CreatedAt,
			UpdatedAt: dmn.UpdatedAt,

			StatusCode: dmn.StatusCode,
			Status:     status,
		})
	}

	return oapi.GetFederationUUIDProject200JSONResponse{
		Count: len(items),
		Items: items,
	}, nil
}

func (a *Web) PatchProjectUUIDName(ctx context.Context, request oapi.PatchProjectUUIDNameRequestObject) (oapi.PatchProjectUUIDNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.ChangeProjectName(request.UUID, request.Body.Name)
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	return oapi.PatchProjectUUIDName200Response{}, nil
}

func (a *Web) PatchProjectUUIDDescription(ctx context.Context, request oapi.PatchProjectUUIDDescriptionRequestObject) (oapi.PatchProjectUUIDDescriptionResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.ChangeProjectDescription(request.UUID, request.Body.Description)
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	return oapi.PatchProjectUUIDDescription200Response{}, nil
}

func (a *Web) PatchProjectUUIDOptions(ctx context.Context, request oapi.PatchProjectUUIDOptionsRequestObject) (oapi.PatchProjectUUIDOptionsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	if request.Body == nil {
		return nil, errors.New("options is nil")
	}

	err := a.app.FederationService.ChangeProjectOptions(request.UUID, domain.ProjectOptions{
		RequireCancelationComment: request.Body.RequireCancelationComment,
		RequireDoneComment:        request.Body.RequireDoneComment,
		StatusEnable:              request.Body.StatusEnable,
		Color:                     request.Body.Color,
	})
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	return oapi.PatchProjectUUIDOptions200Response{}, nil
}

func (a *Web) PatchProjectUUID(ctx context.Context, request oapi.PatchProjectUUIDRequestObject) (oapi.PatchProjectUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	if request.Body == nil {
		return nil, errors.New("options is nil")
	}

	err := a.app.FederationService.ChangeProjectParams(domain.NewCreatorFromUser(&claims), request.UUID, domain.ProjectParams{
		Status:        request.Body.Status,
		StatusSort:    request.Body.StatusSort,
		FieldsSort:    request.Body.FieldsSort,
		ResponsibleBy: request.Body.ResponsibleBy,
	})
	if err != nil {
		return nil, ErrInvalidAuthHeader
	}

	return oapi.PatchProjectUUID200Response{}, nil
}

func (a *Web) PatchProjectUUIDGraph(ctx context.Context, request oapi.PatchProjectUUIDGraphRequestObject) (oapi.PatchProjectUUIDGraphResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	jsonStr, err := json.Marshal(request.Body.Graph)
	if err != nil {
		return nil, err
	}

	if (request.Body.Graph == nil) || (len(request.Body.Graph) == 0) {
		graphMap, err := a.app.FederationService.ChangeProjectStatus(request.UUID, nil)
		if err != nil {
			return nil, err
		}

		return oapi.PatchProjectUUIDGraph200JSONResponse(helpers.ToInterfaceMap(graphMap)), nil
	}

	sg, err := domain.NewStatusGraphFromJSON(string(jsonStr))
	if err != nil {
		return nil, err
	}

	graphMap, err := a.app.FederationService.ChangeProjectStatus(request.UUID, sg)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProjectUUIDGraph200JSONResponse(helpers.ToInterfaceMap(graphMap)), nil
}

func (a *Web) PostProjectUUIDUser(ctx context.Context, request oapi.PostProjectUUIDUserRequestObject) (oapi.PostProjectUUIDUserResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	projectDTO, err := a.app.FederationService.GetProject(request.UUID)
	if err != nil {
		return nil, err
	}

	pu := domain.NewProjectUser(projectDTO.FederationUUID, projectDTO.CompanyUUID, projectDTO.UUID, request.Body.UserUuid)

	err = a.app.FederationService.AddUserToProject(pu)
	if err != nil {
		return nil, err
	}

	return oapi.PostProjectUUIDUser200JSONResponse{
		Uuid: pu.UUID,
	}, nil
}

func (a *Web) DeleteProjectUUIDUserUserUUID(ctx context.Context, request oapi.DeleteProjectUUIDUserUserUUIDRequestObject) (oapi.DeleteProjectUUIDUserUserUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteUserFromProject(request.UUID, request.UserUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProjectUUIDUserUserUUID200Response{}, nil
}

func (a *Web) DeleteProjectUUIDCatalogEntityUUID(ctx context.Context, request oapi.DeleteProjectUUIDCatalogEntityUUIDRequestObject) (oapi.DeleteProjectUUIDCatalogEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteProject(request.UUID.String())
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProjectUUIDCatalogEntityUUID200Response{}, nil
}

func (a *Web) GetCompanyUUIDProjectCatalogEntityName(ctx context.Context, request oapi.GetCompanyUUIDProjectCatalogEntityNameRequestObject) (oapi.GetCompanyUUIDProjectCatalogEntityNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmns, err := a.app.FederationService.GetCompanyProjectCatalogData(request.UUID, request.EntityName)
	if err != nil {
		return nil, err
	}

	return oapi.GetCompanyUUIDProjectCatalogEntityName200JSONResponse{
		Count: len(dmns),
		Items: helpers.Map(dmns, func(item domain.ProjectCatalogData, index int) dto.ProjectCatalogDataDTO {
			project, _ := a.app.DictionaryService.FindProject(item.ProjectUUID)
			projectDTOs := dto.NewProjectDTOs(project)

			return dto.ProjectCatalogDataDTO{
				UUID:      item.UUID,
				Name:      string(item.Name),
				Value:     item.Value,
				Project:   &projectDTOs,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}
		}),
	}, nil
}

func (a *Web) GetProjectUUIDCatalog(ctx context.Context, request oapi.GetProjectUUIDCatalogRequestObject) (oapi.GetProjectUUIDCatalogResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmns, err := a.app.FederationService.GetProjectCatalogData(request.UUID, nil)
	if err != nil {
		return nil, err
	}

	return oapi.GetProjectUUIDCatalog200JSONResponse{
		Count: len(dmns),
		Items: helpers.Map(dmns, func(item domain.ProjectCatalogData, index int) dto.ProjectCatalogDataDTO {
			return dto.ProjectCatalogDataDTO{
				UUID:      item.UUID,
				Name:      string(item.Name),
				Value:     item.Value,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}
		}),
	}, nil
}

func (a *Web) GetProjectUUIDCatalogEntityName(ctx context.Context, request oapi.GetProjectUUIDCatalogEntityNameRequestObject) (oapi.GetProjectUUIDCatalogEntityNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	// @todo: add check for project EntityName existence in allow list
	dmns, err := a.app.FederationService.GetProjectCatalogData(request.UUID, &request.EntityName)
	if err != nil {
		return nil, err
	}

	return oapi.GetProjectUUIDCatalogEntityName200JSONResponse{
		Count: len(dmns),
		Items: helpers.Map(dmns, func(item domain.ProjectCatalogData, index int) dto.ProjectCatalogDataDTO {
			return dto.ProjectCatalogDataDTO{
				UUID:      item.UUID,
				Name:      string(item.Name),
				Value:     item.Value,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}
		}),
	}, nil
}

func (a *Web) PostProjectUUIDCatalog(ctx context.Context, request oapi.PostProjectUUIDCatalogRequestObject) (oapi.PostProjectUUIDCatalogResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	project, found := a.app.DictionaryService.FindProject(request.UUID)
	if !found {
		return nil, dto.NotFoundErr("проект не найден")
	}

	cd := domain.ProjectCatalogData{
		UUID:           uuid.New(),
		FederationUUID: project.FederationUUID,
		CompanyUUID:    project.CompanyUUID,
		ProjectUUID:    project.UUID,
		Name:           request.Body.CatalogName,
		Value:          request.Body.Value,
	}

	err := a.app.FederationService.CreateCatalogData(cd)
	if err != nil {
		return nil, err
	}

	return oapi.PostProjectUUIDCatalog200JSONResponse{
		Uuid: cd.UUID,
	}, nil
}

func (a *Web) PostProjectUUIDFieldEntityUUID(ctx context.Context, request oapi.PostProjectUUIDFieldEntityUUIDRequestObject) (oapi.PostProjectUUIDFieldEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	project, f := a.app.DictionaryService.FindProject(request.UUID)
	if !f {
		return nil, dto.NotFoundErr("проект не найден")
	}

	err := a.app.FederationService.AddProjectField(request.UUID, project.CompanyUUID, request.EntityUUID, request.Body.RequiredOnStatuses, request.Body.Style)
	if err != nil {
		return nil, err
	}

	return oapi.PostProjectUUIDFieldEntityUUID200Response{}, nil
}

func (a *Web) DeleteProjectUUIDFieldEntityUUID(ctx context.Context, request oapi.DeleteProjectUUIDFieldEntityUUIDRequestObject) (oapi.DeleteProjectUUIDFieldEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.RemoveProjectField(request.UUID, request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProjectUUIDFieldEntityUUID200Response{}, nil
}

//

func (a *Web) DeleteProjectUUIDStatusEntityUUID(ctx context.Context, request oapi.DeleteProjectUUIDStatusEntityUUIDRequestObject) (oapi.DeleteProjectUUIDStatusEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.DeleteProjectStatus(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProjectUUIDStatusEntityUUID200Response{}, nil
}

func (a *Web) GetProjectUUIDStatus(ctx context.Context, request oapi.GetProjectUUIDStatusRequestObject) (oapi.GetProjectUUIDStatusResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	cps, err := a.app.FederationService.GetProjectStatuses(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetProjectUUIDStatus200JSONResponse{
		Count: len(cps),
		Items: lo.Map(cps, func(cp domain.ProjectStatus, _ int) dto.ProjectStatusDTO {
			return dto.ProjectStatusDTO{
				UUID:        cp.UUID,
				Name:        cp.Name,
				Color:       cp.Color,
				Number:      cp.Number,
				Description: cp.Description,
				Edit:        cp.Edit,
			}
		}),
	}, nil
}

func (a *Web) PatchProjectUUIDStatusEntityUUID(ctx context.Context, request oapi.PatchProjectUUIDStatusEntityUUIDRequestObject) (oapi.PatchProjectUUIDStatusEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.FederationService.UpdateProjectStatus(request.EntityUUID, request.Body.Name, request.Body.Color, request.Body.Description)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProjectUUIDStatusEntityUUID200Response{}, nil
}

func (a *Web) PostProjectUUIDStatus(ctx context.Context, request oapi.PostProjectUUIDStatusRequestObject) (oapi.PostProjectUUIDStatusResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	project, f := a.app.DictionaryService.FindProject(request.UUID)
	if !f {
		return nil, dto.NotFoundErr("проект не найден")
	}

	dm := domain.ProjectStatus{
		UUID:        helpers.Ptr(uuid.New()),
		CompanyUUID: project.CompanyUUID,
		ProjectUUID: project.UUID,
		Name:        request.Body.Name,
		Number:      request.Body.Number,
		Color:       request.Body.Color,
		Description: request.Body.Description,
	}

	err := a.app.FederationService.CreateProjectStatus(dm)
	if err != nil {
		return nil, err
	}

	return oapi.PostProjectUUIDStatus200JSONResponse{
		Uuid: *dm.UUID,
	}, nil
}
