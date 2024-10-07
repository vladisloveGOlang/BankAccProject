package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	UUID           uuid.UUID `json:"uuid"`
	Name           string    `json:"name"`
	Color          string    `json:"color"`
	FederationUUID uuid.UUID `json:"federation_uuid"`
	CompanyUUID    uuid.UUID `json:"company_uuid"`

	CreatedBy User `json:"created_by"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
