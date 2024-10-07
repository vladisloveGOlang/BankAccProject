package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type IDict interface {
	FindUser(email string) (*UserDTO, bool)
	FindUserByUUID(uid uuid.UUID) (*UserDTO, bool)
	FindUsers(emails []string) ([]UserDTO, []string)
	FindTag(uuid uuid.UUID) (*TagDTO, bool)
	FindFederation(uuid uuid.UUID) (*FederationDTO, bool)
	FindProject(uuid uuid.UUID) (*ProjectDTO, bool)
	FindCompanyFields(uuid uuid.UUID) ([]CompanyFieldDTO, bool)
	FindProjectFields(uuid uuid.UUID) ([]ProjectFieldDTO, bool)
	FindCatalogFields(uuid uuid.UUID) ([]CatalogFieldDTO, bool)
	FindCompanyPriorities(uuid.UUID, int) (*CompanyPriorityDTO, bool)
}

type OpenByDTO struct {
	OpenAt time.Time `json:"open_at"`
	OpenBy UserDTO   `json:"open_by"`
}

type TaskDTO struct {
	UUID          uuid.UUID `json:"uuid"`
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CreatedBy     UserDTO   `json:"created_by"`
	ResponsibleBy *UserDTO  `json:"responsible_by,omitempty"`
	ImplementBy   *UserDTO  `json:"implement_by,omitempty"`
	ManagedBy     *UserDTO  `json:"managed_by,omitempty"`

	IsEpic bool `json:"is_epic"`

	CoWorkersBy []UserDTO `json:"co_workers_by"`
	WatchBy     []UserDTO `json:"watch_by"`
	Tags        []string  `json:"tags"`
	CompanyTags []TagDTOs `json:"company_tags"`

	Federation FederationDTOs `json:"federation"`
	Project    ProjectDTOs    `json:"project"`

	Path []string `json:"path"`

	Icon            string             `json:"icon"`
	Status          StatusDTO          `json:"status"`
	Priority        int                `json:"priority"`
	CompanyPriority CompanyPriorityDTO `json:"company_priority"`

	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
	FinishedBy *UserDTO   `json:"finished_by,omitempty"`
	FinishTo   *time.Time `json:"finish_to"`
	Duration   int        `json:"duration"`

	UpdatedAt  time.Time  `json:"updated_at"`
	ActivityAt time.Time  `json:"activity_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	Fields []TaskFieldDTO `json:"fields"`

	Comments []CommentDTO `json:"comments"`
	Files    []FileDTOs   `json:"files"`

	Reminders []ReminderDTO `json:"reminders"`

	CommentsTotal  int         `json:"comments_total"`
	ChildrensTotal int         `json:"childrens_total"`
	ChildrensUUID  []uuid.UUID `json:"childrens_uuid"`

	// @todo: renaim
	LinkedFieldsData map[uuid.UUID]interface{} `json:"linked_fields_data"`

	Stops []domain.Stop `json:"stops"`

	IsLiked *bool `json:"is_liked,omitempty"`

	FirstOpen []OpenByDTO `json:"first_open"`
	Views     int         `json:"views"`

	Activities Pagination[ActivityDTO] `json:"activities"`
}

type Pagination[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Count int64 `json:"count"`
}

type TaskDTOs struct {
	UUID          uuid.UUID `json:"uuid" xlsx:"A"`
	ID            int       `json:"id"`
	Name          string    `json:"name" xlsx:"B" ru:"Название"`
	CreatedBy     UserDTO   `json:"created_by" xlsx:"C" ru:"Создан(а)"`
	ResponsibleBy *UserDTO  `json:"responsible_by,omitempty"`
	ImplementBy   *UserDTO  `json:"implement_by,omitempty"`
	ManagedBy     *UserDTO  `json:"managed_by,omitempty"`

	IsEpic bool `json:"is_epic"`

	CoWorkersBy []UserDTO `json:"co_workers_by"`
	WatchBy     []UserDTO `json:"watch_by"`
	Tags        []string  `json:"tags" xlsx:"J" ru:"Теги"`

	Federation FederationDTOs `json:"federation"`
	Project    ProjectDTOs    `json:"project"`

	Path []string `json:"path"`

	Status   StatusDTO `json:"status" xlsx:"D" ru:"Статус"`
	Priority int       `json:"priority" xlsx:"E" ru:"Приоритет"`

	CreatedAt  time.Time  `json:"created_at" xlsx:"F" ru:"Создано"`
	FinishedAt *time.Time `json:"finished_at,omitempty" xlsx:"G" ru:"Завершено"`
	FinishedBy *UserDTO   `json:"finished_by,omitempty"`

	FinishTo *time.Time `json:"finish_to,omitempty"`
	Duration int        `json:"duration"`

	UpdatedAt  time.Time  `json:"updated_at" xlsx:"H" ru:"Обновлено"`
	ActivityAt time.Time  `json:"activity_at" xlsx:"I" ru:"Активность"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" xlsx:"J" ru:"Удалено"`

	ChildrensTotal int `json:"childrens_total"  xlsx:"J" ru:"Потомков"`
}

type TaskFieldDTO struct {
	Hash     string      `json:"hash"`
	Name     string      `json:"name"`
	DataType int         `json:"data_type"`
	Value    interface{} `json:"value"`
}

func NewTaskDTO(dm domain.Task, comments []domain.Comment, files []domain.File, reminders []domain.Reminder, linkedFieldsData map[uuid.UUID]interface{}, dict IDict, s3 IStorage) TaskDTO {
	createdBy, _ := dict.FindUser(dm.CreatedBy)
	implementBy, fi := dict.FindUser(dm.ImplementBy)

	responsibleBy, fr := dict.FindUser(dm.ResponsibleBy)
	finishedBy, fb := dict.FindUser(dm.FinishedBy)
	managedBy, fm := dict.FindUser(dm.ManagedBy)
	federationDTO, _ := dict.FindFederation(dm.FederationUUID)
	projectDTO, _ := dict.FindProject(dm.ProjectUUID)

	coWorkers, _ := dict.FindUsers(dm.CoWorkersBy)
	watchBy, _ := dict.FindUsers(dm.WatchBy)

	// todo: add company to task dto and table
	project, f := dict.FindProject(dm.ProjectUUID)
	if !f {
		logrus.Errorf("project not found: %s", dm.ProjectUUID)
	}

	projectFields, _ := dict.FindProjectFields(project.UUID)

	taskFields := []TaskFieldDTO{}

	if dm.Fields != nil {
		fi := dm.Fields
		for _, pf := range projectFields {
			if _, ok := fi[pf.Hash]; ok {
				taskFields = append(taskFields, TaskFieldDTO{
					Hash:     pf.Hash,
					Name:     pf.Name,
					DataType: pf.DataType,
					Value:    fi[pf.Hash],
				})
			}
		}
	}

	// comments
	commentsDtos := []CommentDTO{}
	for _, dm := range comments {
		commentsDtos = append(commentsDtos, NewCommentDTO(dm, dict, s3))
	}

	// comments
	remindersDtos := lo.Map(reminders, func(dm domain.Reminder, index int) ReminderDTO {
		return ReminderDTO{
			UUID:        dm.UUID,
			TaskUUID:    dm.TaskUUID,
			DateFrom:    dm.DateFrom,
			Description: dm.Description,
			Type:        dm.Type,
			Comment:     dm.Comment,
			DateTo:      dm.DateTo,
			CreatedAt:   dm.CreatedAt,
			UpdatedAt:   dm.UpdatedAt,
		}
	})

	// first open
	firstOpen := []OpenByDTO{}

	if len(dm.FirstOpen) > 0 {
		for userUUID, fo := range dm.FirstOpen {
			userUID, err := uuid.Parse(userUUID)
			if err != nil {
				logrus.Errorf("first open user uuid parse error: %s", userUUID)
				continue
			}

			user, f := dict.FindUserByUUID(userUID)
			if f {
				firstOpen = append(firstOpen, OpenByDTO{
					OpenAt: fo,
					OpenBy: *user,
				})
			} else {
				logrus.Errorf("first open user not found: %s", userUUID)
			}
		}
	}

	// files
	filesDtos := lo.Map(files, func(dm domain.File, index int) FileDTOs {
		createdBy, f := dict.FindUserByUUID(dm.CreatedBy)
		if !f {
			logrus.Errorf("file created by not found: %s", dm.CreatedBy)
		}

		return FileDTOs{
			UUID:      dm.UUID,
			Name:      dm.Name,
			Ext:       dm.Ext,
			Size:      dm.Size,
			URL:       dm.URL,
			CreatedAt: dm.CreatedAt,
			CreatedBy: *createdBy,
		}
	})

	// Tags
	companyTags := []TagDTOs{}
	for _, tag := range dm.Tags {
		tagUID, f := uuid.Parse(tag)
		if f == nil {
			tagDTO, f := dict.FindTag(tagUID)
			if f {
				companyTags = append(companyTags, TagDTOs{
					UUID:  tagDTO.UUID,
					Name:  tagDTO.Name,
					Color: tagDTO.Color,
				})
			} else {
				logrus.Errorf("tag not found: %s", tagUID)
			}
		}
	}

	tags := lo.Filter(dm.Tags, func(t string, _ int) bool {
		_, f := uuid.Parse(t)
		return f != nil
	})

	//

	companyPriority := CompanyPriorityDTO{
		Number: dm.Priority,
		Color:  "#000000",
		Name:   "",
	}

	if dm.Priority >= 10 {
		cp, f := dict.FindCompanyPriorities(project.CompanyUUID, dm.Priority)
		if f {
			companyPriority = *cp
		} else {
			logrus.Errorf("company priority not found: %s", dm.CompanyUUID)
		}
	}

	return TaskDTO{
		UUID:       dm.UUID,
		Name:       dm.Name,
		ID:         dm.ID,
		Project:    NewProjectDTOs(projectDTO),
		Federation: NewFederationDTOs(federationDTO),

		Description: dm.Description,

		Priority:        dm.Priority,
		CompanyPriority: companyPriority,
		IsEpic:          dm.IsEpic,
		Icon:            dm.Icon,

		Path: dm.Path,

		CreatedBy:     *createdBy,
		CoWorkersBy:   coWorkers,
		WatchBy:       watchBy,
		ManagedBy:     helpers.Empty(*managedBy, fm),
		ResponsibleBy: helpers.Empty(*responsibleBy, fr),
		ImplementBy:   helpers.Empty(*implementBy, fi),
		Tags:          tags,
		CompanyTags:   companyTags,
		Status: StatusDTO{
			Code: dm.Status,
			Name: "todo",
		},
		Fields: taskFields,

		CreatedAt:  dm.CreatedAt,
		UpdatedAt:  dm.UpdatedAt,
		ActivityAt: dm.ActivityAt,
		DeletedAt:  dm.DeletedAt,

		Comments:  commentsDtos,
		Files:     filesDtos,
		Reminders: remindersDtos,

		CommentsTotal:  dm.CommentsTotal,
		ChildrensTotal: dm.ChildrensTotal,
		ChildrensUUID:  dm.ChildrensUUID,

		LinkedFieldsData: linkedFieldsData,

		Stops: dm.Stops,

		FinishTo:   dm.FinishTo,
		FinishedAt: dm.FinishedAt,
		FinishedBy: helpers.Empty(*finishedBy, fb),

		FirstOpen: firstOpen,
		Views:     len(firstOpen),

		Activities: Pagination[ActivityDTO]{
			Items: lo.Map(dm.Activities, func(dm domain.Activity, _ int) ActivityDTO {
				createdBy, f := dict.FindUser(dm.CreatedBy.Email)

				if !f {
					logrus.Errorf("activity created by not found: %s", dm.CreatedBy.Email)
				}

				return *NewActivityDTO(dm, *createdBy)
			}),
			Total: dm.ActivitiesTotal,
			Count: int64(len(dm.Activities)),
		},
	}
}

func NewTaskDTOs(dm domain.Task, dict IDict) TaskDTOs {
	createdBy, _ := dict.FindUser(dm.CreatedBy)
	implementBy, fi := dict.FindUser(dm.ImplementBy)
	managedBy, _ := dict.FindUser(dm.ManagedBy)

	responsibleBy, fr := dict.FindUser(dm.ResponsibleBy)
	federationDTO, _ := dict.FindFederation(dm.FederationUUID)
	projectDTO, _ := dict.FindProject(dm.ProjectUUID)

	coWorkers, _ := dict.FindUsers(dm.CoWorkersBy)
	watchBy, _ := dict.FindUsers(dm.WatchBy)

	return TaskDTOs{
		UUID:       dm.UUID,
		Name:       dm.Name,
		ID:         dm.ID,
		Project:    NewProjectDTOs(projectDTO),
		Federation: NewFederationDTOs(federationDTO),

		Priority: dm.Priority,
		IsEpic:   dm.IsEpic,

		Path: dm.Path,

		CreatedBy:     *createdBy,
		CoWorkersBy:   coWorkers,
		WatchBy:       watchBy,
		ManagedBy:     managedBy,
		ResponsibleBy: helpers.Empty(*responsibleBy, fr),
		ImplementBy:   helpers.Empty(*implementBy, fi),
		Tags:          dm.Tags,
		Status: StatusDTO{
			Code: dm.Status,
			Name: "todo",
		},

		ChildrensTotal: dm.ChildrensTotal,
		FinishedAt:     dm.FinishedAt,
		FinishTo:       dm.FinishTo,

		CreatedAt:  dm.CreatedAt,
		ActivityAt: dm.ActivityAt,
		UpdatedAt:  dm.UpdatedAt,
		DeletedAt:  dm.DeletedAt,
	}
}

type FilterDTO struct {
	Name     string
	Operator string
	Value    interface{}
}

func NewFilterDTO(filter *string) (dtos []FilterDTO, err error) {
	if filter == nil {
		return dtos, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(*filter), &jsonMap)
	if err != nil {
		return dtos, err
	}

	for k, v := range jsonMap {
		dtos = append(dtos, FilterDTO{
			Name:     k,
			Operator: "=",
			Value:    fmt.Sprintf("%v", v),
		})
	}

	return dtos, err
}

type TaskSearchDTO struct {
	MyEmail *string `json:"my_email"`

	Name           *string   `json:"name"`
	Offset         *int      `json:"offset"`
	Limit          *int      `json:"limit"`
	IsMy           *bool     `json:"is_my"`
	Status         *int      `json:"status"`
	IsEpic         *bool     `json:"is_epic"`
	ProjectUUID    uuid.UUID `json:"project_uuid"`
	FederationUUID uuid.UUID `json:"federation_uuid"`
	Participated   *[]string `json:"participated"`
	Tags           *[]string `json:"tags"`
	Path           *string   `json:"path"`

	Fields []FilterDTO `json:"fields"`

	Order *string `json:"order"`
	By    *string `json:"by"`
}

func (d *TaskSearchDTO) Validate() error {
	if d.FederationUUID == uuid.Nil {
		return errors.New("federation_uuid не может быть пустым")
	}

	if d.ProjectUUID == uuid.Nil {
		return errors.New("project_uuid не может быть пустым")
	}

	return nil
}
