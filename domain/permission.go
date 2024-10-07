package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	UUID           uuid.UUID `json:"uuid"`
	UserUUID       uuid.UUID `json:"user_uuid"`
	FederationUUID uuid.UUID `json:"federation_uuid"`

	Rules PermissionRules `json:"rules"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type PermissionRules struct {
	FederationPatch      bool `json:"federation_patch"`
	FederationInviteUser bool `json:"federation_invite_user"`
	FederationDeleteUser bool `json:"federation_delete_user"`
	CompanyCreate        bool `json:"company_create"`
	CompanyDelete        bool `json:"company_delete"`
	CompanyPatch         bool `json:"company_patch"`
	CompanyAddUser       bool `json:"company_add_user"`
	CompanyDeleteUser    bool `json:"company_delete_user"`
	ProjectCreate        bool `json:"project_create"`
	ProjectDelete        bool `json:"project_delete"`
	ProjectPatch         bool `json:"project_patch"`
	ProjectAddUser       bool `json:"project_add_user"`
	ProjectDeleteUser    bool `json:"project_delete_user"`
	TaskCreate           bool `json:"task_create"`
	TaskDelete           bool `json:"task_delete"`
	TaskPatch            bool `json:"task_patch"`
}

func (j *PermissionRules) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := PermissionRules{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j PermissionRules) Value() (driver.Value, error) {
	return json.Marshal(j)
}
