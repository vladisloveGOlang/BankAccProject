package dto

import (
	"time"

	"github.com/google/uuid"
)

type CompanyDTO struct {
	UUID           uuid.UUID `json:"uuid"`
	Name           string    `json:"name"`
	FederationUUID uuid.UUID `json:"federation_uuid"`

	UsersTotal int       `json:"users_total"`
	Users      []UserDTO `json:"users,omitempty"`

	ProjectsTotal int                  `json:"projects_total"`
	Projects      []ProjectDTOs        `json:"projects,omitempty"`
	Priorities    []CompanyPriorityDTO `json:"priorities,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CompanyDTOs struct {
	UUID           uuid.UUID  `json:"uuid"`
	Name           string     `json:"name"`
	IsLiked        *bool      `json:"is_liked,omitempty"`
	FederationUUID *uuid.UUID `json:"federation_uuid,omitempty"`
}

type UserCompanyDTOs struct {
	Company  CompanyDTOs `json:"company"`
	Position string      `json:"position"`
}

type CompanyPriorityDTO struct {
	UUID   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	Number int       `json:"priority"`
	Color  string    `json:"color"`
}
