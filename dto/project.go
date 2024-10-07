package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProjectDTO struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`

	CompanyUUID    uuid.UUID `json:"company_uuid,omitempty"`
	FederationUUID uuid.UUID `json:"federation_uuid,omitempty"`

	Company    CompanyDTOs    `json:"company"`
	Federation FederationDTOs `json:"federation"`

	ResponsibleBy *UserDTO `json:"responsible_by,omitempty"`

	Fields      []ProjectFieldDTO `json:"fields"`
	FieldsTotal int               `json:"fields_total"`

	StatusGraph *map[string][]string `json:"status_graph,omitempty"`

	Options *ProjectOptionsDTO `json:"options,omitempty"`

	Users []ProjectUserDto `json:"users"`

	AllowSort []string `json:"allow_sort"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StatusCode      int              `json:"status_code,omitempty"`
	Status          ProjectStatusDTO `json:"status"`
	StatusUpdatedAt *time.Time       `json:"status_updated_at,omitempty"`

	Statuses *[]ProjectStatusDTO `json:"statuses,omitempty"`

	Statistic       *ProjectStatistics `json:"statistic,omitempty"`
	FieldStatistics []FieldStatistics  `json:"field_statistics,omitempty"`
}

type ProjectStatistics struct {
	TasksTotal         int `json:"tasks_total"`
	TasksFinishedTotal int `json:"tasks_finished_total"`
	TasksActiveTotal   int `json:"tasks_active_total"`
	TaskCanceledTotal  int `json:"task_canceled_total"`
	TaskDeletedTotal   int `json:"task_deleted_total"`
}

type FieldStatistics struct {
	Name   string  `json:"name"`
	Filled float64 `json:"filled"`
	Count  int     `json:"count"`
	Total  int     `json:"total"`
}

type ProjectOptionsDTO struct {
	RequireCancelationComment *bool   `json:"require_cancelation_comment"`
	RequireDoneComment        *bool   `json:"require_done_comment"`
	StatusEnable              *bool   `json:"status_enable"`
	Color                     *string `json:"color"`
}

type ProjectDTOs struct {
	UUID           uuid.UUID        `json:"uuid"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	IsLiked        *bool            `json:"is_liked,omitempty"`
	Status         ProjectStatusDTO `json:"status"`
	StatusCode     int              `json:"status_code"`
	FederationUUID uuid.UUID        `json:"federation_uuid"`
	CompanyUUID    uuid.UUID        `json:"company_uuid"`
}

func NewProjectDTOs(dto *ProjectDTO) ProjectDTOs {
	if dto == nil {
		return ProjectDTOs{}
	}

	return ProjectDTOs{
		UUID:           dto.UUID,
		Name:           dto.Name,
		Description:    dto.Description,
		StatusCode:     dto.StatusCode,
		FederationUUID: dto.FederationUUID,
		CompanyUUID:    dto.CompanyUUID,
	}
}

type CompanyFieldDTO struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Hash        string    `json:"hash"`
	DataType    int       `json:"data_type"`
	DataDesc    string    `json:"data_desc"`

	ProjectsUUID      []uuid.UUID `json:"project_uuids"`
	TasksTotal        int         `json:"tasks_total"`
	TasksFilled       int         `json:"tasks_filled"`
	TasksActiveFilled int         `json:"tasks_active_filled"`
	TasksActiveTotal  int         `json:"tasks_active_total"`
}

type ProjectFieldDTO struct {
	UUID               uuid.UUID `json:"uuid"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Hash               string    `json:"hash"`
	DataType           int       `json:"data_type"`
	DataDesc           string    `json:"data_desc"`
	RequiredOnStatuses []int     `json:"required_on_statuses"`
	Style              string    `json:"style"`

	ProjectUUID uuid.UUID `json:"project_uuid"`
}

type ProjectUserDto struct {
	UUID    uuid.UUID `json:"uuid"`
	User    UserDTO   `json:"user"`
	AddedAt string    `json:"added_at"`
}

type ProjectCatalogDataDTO struct {
	UUID  uuid.UUID `json:"uuid"`
	Name  string    `json:"name"`
	Value string    `json:"value"`

	Project *ProjectDTOs `json:"project,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectStatusDTO struct {
	UUID        *uuid.UUID `json:"uuid,omitempty"`
	Name        string     `json:"name"`
	Number      int        `json:"number"`
	Color       string     `json:"color"`
	Description string     `json:"description"`
	Edit        []string   `json:"edit"`
}

type ProjectStatusDTOs struct {
	Name        string  `json:"name"`
	Number      int     `json:"number"`
	Color       string  `json:"color"`
	Description *string `json:"description,omitempty"`
}

func (s *ProjectStatusDTO) ToDTOs() ProjectStatusDTOs {
	return ProjectStatusDTOs{
		Name:        s.Name,
		Number:      s.Number,
		Color:       s.Color,
		Description: &s.Description,
	}
}
