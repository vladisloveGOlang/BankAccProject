package dto

import (
	"time"

	"github.com/google/uuid"
)

type PermissionDTO struct {
	UUID           uuid.UUID  `json:"uuid"`
	UserUUID       uuid.UUID  `json:"user_uuid"`
	FederationUUID uuid.UUID  `json:"federation_uuid"`
	CompanyUUID    *uuid.UUID `json:"company_uuid"`
	ProjectUUID    *uuid.UUID `json:"project_uuid"`

	Rules PermissionRulesDTO `json:"rules"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type PermissionRulesDTO struct {
	FederationPatch      bool `json:"federation_patch"`
	FederationInviteUser bool `json:"federation_invite_user"`
	FederationDeleteUser bool `json:"federation_delete_user"`

	CompanyCreate     bool `json:"company_create"`
	CompanyDelete     bool `json:"company_delete"`
	CompanyPatch      bool `json:"company_patch"`
	CompanyAddUser    bool `json:"company_add_user"`
	CompanyDeleteUser bool `json:"company_delete_user"`

	ProjectCreate     bool `json:"project_create"`
	ProjectDelete     bool `json:"project_delete"`
	ProjectPatch      bool `json:"project_patch"`
	ProjectAddUser    bool `json:"project_add_user"`
	ProjectDeleteUser bool `json:"project_delete_user"`

	TaskCreate bool `json:"task_create"`
	TaskDelete bool `json:"task_delete"`
	TaskPatch  bool `json:"task_patch"`
}
