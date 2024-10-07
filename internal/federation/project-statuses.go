package federation

import (
	"sort"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

func (s *Service) CreateProjectStatus(dm domain.ProjectStatus) (err error) {
	err = s.repo.CreateProjectStatus(dm)
	return err
}

func (s *Service) GetProjectStatus(uid uuid.UUID) (dm domain.ProjectStatus, err error) {
	orm, err := s.repo.GetProjectStatus(uid)
	if err != nil {
		return dm, err
	}

	return domain.ProjectStatus{
		UUID:        orm.UUID,
		Name:        orm.Name,
		Number:      orm.Number,
		Color:       orm.Color,
		CompanyUUID: orm.CompanyUUID,
	}, err
}

func (s *Service) GetProjectStatuses(projectUUID uuid.UUID) (tags []domain.ProjectStatus, err error) {
	pr, err := s.repo.GetProject(projectUUID)
	if err != nil {
		return tags, err
	}

	orms, err := s.repo.GetProjectStatuses(pr.UUID)
	if err != nil {
		return tags, err
	}

	tags = make([]domain.ProjectStatus, 0)

	for _, orm := range orms {
		edit := []string{"name", "color"}
		n := orm.Number
		color := orm.Color
		name := orm.Name

		if n == domain.StatusUnknown {
			edit = []string{}
			name = "Необработана"
		}

		if n == domain.StatusDone {
			edit = []string{"name"}
		}

		if n == domain.StatusCancel {
			edit = []string{"name"}
		}

		tags = append(tags, domain.ProjectStatus{
			UUID:        orm.UUID,
			Name:        name,
			Description: orm.Description,
			Number:      orm.Number,
			Color:       color,
			CompanyUUID: pr.CompanyUUID,
			ProjectUUID: pr.UUID,
			Edit:        edit,
		})
	}
	tagsNumbers := lo.Map(tags, func(tag domain.ProjectStatus, _ int) int {
		return tag.Number
	})

	for n, tag := range domain.GetTaskStatuses() {
		if lo.IndexOf(tagsNumbers, n) >= 0 {
			continue
		}

		edit := []string{"name", "color"}
		if n == domain.StatusUnknown {
			edit = []string{}
		}

		if n == domain.StatusDone {
			edit = []string{"name"}
		}

		if n == domain.StatusCancel {
			edit = []string{"name"}
		}

		tags = append(tags, domain.ProjectStatus{
			UUID:        nil,
			Name:        tag,
			Number:      n,
			Color:       "",
			Description: pr.Description,
			CompanyUUID: pr.CompanyUUID,
			ProjectUUID: pr.UUID,
			Edit:        edit,
		})
	}

	lo.Reverse(pr.StatusSort)

	// Sort and mv Unknown status to the first position
	pr.StatusSort = lo.Filter(pr.StatusSort, func(n int, _ int) bool {
		return n != domain.StatusUnknown
	})
	pr.StatusSort = append(pr.StatusSort, domain.StatusUnknown)
	sort.Slice(tags, func(i, j int) bool {
		return lo.IndexOf(pr.StatusSort, tags[i].Number) > lo.IndexOf(pr.StatusSort, tags[j].Number)
	})

	return tags, err
}

func (s *Service) UpdateProjectStatus(uid uuid.UUID, name, color, description string) (err error) {
	err = s.repo.UpdateProjectStatus(uid, name, color, description)
	return err
}

func (s *Service) DeleteProjectStatus(uid uuid.UUID) (err error) {
	err = s.repo.DeleteProjectStatus(uid)
	return err
}
