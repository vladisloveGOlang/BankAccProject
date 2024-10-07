package dto

import (
	"time"

	"github.com/google/uuid"
)

type SmsDTO struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	UserUUID       uuid.UUID

	Phone  string
	Text   string
	Status string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type SmsFilterDTO struct {
	CompanyUUID *uuid.UUID `json:"company_uuid"`
	Offset      *int       `json:"offset"`
	Limit       *int       `json:"limit"`
	IsMy        *bool      `json:"is_my"`
	Status      *int       `json:"status"`
	MyEmail     *string    `json:"my_email"`
}
