package dto

import (
	"time"

	"github.com/google/uuid"
)

type AgentDTO struct {
	UUID uuid.UUID `json:"uuid"`

	Name string `json:"name"`

	FederationUUID uuid.UUID  `json:"federation_uuid"`
	CompanyUUID    *uuid.UUID `json:"company_uuid,omitempty"`
	ProjectUUID    *uuid.UUID `json:"project_uuid,omitempty"`

	Contacts []AgentContactsDTO `json:"contacts"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgentContactsDTO struct {
	Type string `json:"type"`
	Val  string `json:"val"`
}
