package dto

import (
	"time"

	"github.com/google/uuid"
)

type InviteDTO struct {
	UUID uuid.UUID `json:"uuid"`

	Email          string     `json:"email"`
	FederationUUID uuid.UUID  `json:"federation_uuid"`
	CompanyUUID    *uuid.UUID `json:"company_uuid"`

	Federation *FederationDTOs `json:"federation,omitempty"`
	Company    *CompanyDTOs    `json:"company,omitempty"`

	AcceptedAt *time.Time `json:"accepted_at"`
	DeclinedAt *time.Time `json:"declined_at"`

	CreatedAt time.Time `json:"created_at"`
}
