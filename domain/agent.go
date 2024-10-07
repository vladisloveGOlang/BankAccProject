package domain

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    *uuid.UUID
	ProjectUUID    *uuid.UUID
	CreatedBy      string
	CreatedByUUID  uuid.UUID

	Name     string
	Contacts []AgentContacts

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type AgentContacts struct {
	Type string `json:"type"`
	Val  string `json:"val"`
}

type AgentFilter struct {
	FederationUUID uuid.UUID  `json:"federation_uuid"`
	CompanyUUID    *uuid.UUID `json:"company_uuid"`
	Offset         *int       `json:"offset"`
	Limit          *int       `json:"limit"`
	Name           *string    `json:"name"`
}

func NewAgent(federationUUID uuid.UUID, companyUUID *uuid.UUID, me Me, name string, contacts []AgentContacts) *Agent {
	return &Agent{
		UUID:           uuid.New(),
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		CreatedBy:      me.Email,
		CreatedByUUID:  me.UUID,
		Name:           name,
		Contacts:       contacts,
	}
}
