package federation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
)

func (s *Service) CreateProgect(project *domain.Project) (err error) {
	errs, ok := helpers.ValidationStruct(project)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return err
	}

	err = s.repo.CreateProject(project)

	return err
}

func (s *Service) GetProject(uid uuid.UUID) (item domain.Project, err error) {
	orm, err := s.repo.GetProject(uid)
	if err != nil {
		return item, err
	}

	fileds, err := s.GetProjectFields(orm.UUID)
	if err != nil {
		return item, err
	}

	users, err := s.repo.GetProjectUsers(orm.UUID)
	if err != nil {
		return item, err
	}

	sg, err := domain.NewStatusGraphFromJSON(orm.StatusGraph)
	if err != nil {
		return item, err
	}

	item = domain.Project{
		UUID: orm.UUID,

		FederationUUID: orm.FederationUUID,
		CompanyUUID:    orm.CompanyUUID,

		Name:        orm.Name,
		Description: orm.Description,
		Fields:      fileds,
		Meta:        orm.Meta,
		StatusGraph: sg,

		ResponsibleBy: orm.ResponsibleBy,

		Options: orm.Options,

		Users: users,

		CreatedAt: orm.CreatedAt,
		UpdatedAt: orm.UpdatedAt,

		StatusCode:      orm.Status,
		StatusUpdatedAt: orm.StatusUpdatedAt,
	}

	return item, err
}

func (s *Service) GetProjects(federationUUID uuid.UUID) (item []domain.Project, err error) {
	orms, err := s.repo.GetProjects(federationUUID)
	if err != nil {
		return item, err
	}

	for _, orm := range orms {
		users, err := s.repo.GetProjectUsers(orm.UUID)
		if err != nil {
			return item, err
		}

		sg, err := domain.NewStatusGraphFromJSON(orm.StatusGraph)
		if err != nil {
			return item, err
		}

		item = append(item, domain.Project{
			UUID: orm.UUID,

			FederationUUID: orm.FederationUUID,
			CompanyUUID:    orm.CompanyUUID,

			Name:        orm.Name,
			Description: orm.Description,
			Meta:        orm.Meta,
			StatusGraph: sg,

			ResponsibleBy: orm.ResponsibleBy,

			Options: orm.Options,

			Users: users,

			CreatedAt: orm.CreatedAt,
			UpdatedAt: orm.UpdatedAt,

			StatusCode:      orm.Status,
			StatusUpdatedAt: orm.StatusUpdatedAt,
		})
	}

	return item, err
}

func (s *Service) GetProjectStatistic(companyUID, uid uuid.UUID) (item ProjectStatistic, dtos []dto.FieldStatistics, err error) {
	orm, fs, err := s.repo.ProjectStatistic(companyUID, uid)
	if err != nil {
		return item, dtos, err
	}

	return orm, lo.Map(fs, func(item FieldStatistics, _ int) dto.FieldStatistics {
		return dto.FieldStatistics{
			Name:   item.Name,
			Filled: item.Filled,
			Count:  item.Count,
			Total:  item.Total,
		}
	}), err
}

func (s *Service) GetProjectsByUser(ctx context.Context, uid uuid.UUID) (dmns []domain.Project, err error) {
	dmns, err = s.repo.GetProjectsByUser(ctx, uid)
	if err != nil {
		return dmns, err
	}

	return dmns, err
}

func (s *Service) ChangeProjectName(uid uuid.UUID, name string) (err error) {
	p := domain.NewProjectUUID(uid)
	err = p.ChangeName(name)
	if err != nil {
		return err
	}

	err = s.repo.ChangeProjectField(p.UUID, "name", p.Name)

	return err
}

func (s *Service) ChangeProjectDescription(uid uuid.UUID, description string) (err error) {
	p := domain.NewProjectUUID(uid)
	err = p.ChangeDescription(description)
	if err != nil {
		return err
	}

	err = s.repo.ChangeProjectField(p.UUID, "description", p.Description)

	return err
}

func (s *Service) ChangeProjectOptions(uid uuid.UUID, options domain.ProjectOptions) (err error) {
	j, err := json.Marshal(options)
	if err != nil {
		return err
	}

	err = s.repo.gorm.DB.
		Exec("UPDATE projects SET updated_at = NOW(), options = options || ? WHERE uuid = ?", j, uid).
		Error

	if err == nil {
		s.repo.PubUpdate()
	}

	return err
}

func (s *Service) ChangeProjectParams(crtr domain.Creator, uid uuid.UUID, options domain.ProjectParams) (err error) {
	if options.Status != nil {
		err = s.repo.ChangeProjectField(uid, "status", *options.Status)
		if err != nil {
			return err
		}

		stop := Stop{
			CreatedAt:     time.Now(),
			StatusID:      *options.Status,
			CreatedByUUID: crtr.UUID,
		}

		err = s.repo.gorm.DB.Exec("UPDATE projects SET stops = stops::jsonb || ?  WHERE uuid = ?", stop, uid).Error
		if err != nil {
			return err
		}
	}

	if options.ResponsibleBy != nil {
		err = s.repo.ChangeProjectField(uid, "responsible_by", *options.ResponsibleBy)
		if err != nil {
			return err
		}
	}

	if options.StatusSort != nil {
		var IDs []string
		for _, i := range *options.StatusSort {
			IDs = append(IDs, strconv.Itoa(i))
		}

		value := fmt.Sprintf("[%s]", strings.Join(IDs, ","))
		err = s.repo.ChangeProjectField(uid, "status_sort", value)
		if err != nil {
			return err
		}
	}

	if options.FieldsSort != nil {
		value := fmt.Sprintf("[%s]", strings.Join(*options.FieldsSort, ","))
		err = s.repo.ChangeProjectField(uid, "fields_sort", value)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) ChangeProjectStatus(uid uuid.UUID, sg *domain.StatusGraph) (mp map[string][]string, err error) {
	if sg == nil {
		err = s.repo.ChangeProjectField(uid, "status_graph", "{}")
		return make(map[string][]string), err
	}

	err = s.repo.ChangeProjectField(uid, "status_graph", sg.Graph)

	return sg.Graph, err
}

func (s *Service) AddUserToProject(fu *domain.ProjectUser) (err error) {
	err = s.repo.AddUserToProject(fu)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteUserFromProject(projectUUID, userUUID uuid.UUID) (err error) {
	err = s.repo.DeleteUserFromProject(projectUUID, userUUID)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) CreateCatalogData(cd domain.ProjectCatalogData) (err error) {
	err = s.repo.CreateCatalogData(cd)

	return err
}

func (s *Service) DeleteCatalogData(uid uuid.UUID) (err error) {
	err = s.repo.DeleteCatalogData(uid)

	return err
}

func (s *Service) GetProjectCatalogData(uid uuid.UUID, catalogName *string) ([]domain.ProjectCatalogData, error) {
	return s.repo.GetProjectCatalogData(uid, catalogName)
}

func (s *Service) GetCompanyProjectCatalogData(companyUUID uuid.UUID, catalogName string) ([]domain.ProjectCatalogData, error) {
	return s.repo.GetCompanyProjectCatalogData(companyUUID, catalogName)
}

func (s *Service) AddProjectField(uid, companyUUID, companyFieldUUID uuid.UUID, requiredOnstatuses []int, style string) error {
	err := s.repo.AddProjectField(uid, companyUUID, companyFieldUUID, requiredOnstatuses, style)

	if err == nil {
		s.repo.PubUpdate()
	}

	return err
}

func (s *Service) RemoveProjectField(uid, companyFieldUUID uuid.UUID) error {
	err := s.repo.RemoveProjectField(uid, companyFieldUUID)

	if err == nil {
		s.repo.PubUpdate()
	}

	return err
}
