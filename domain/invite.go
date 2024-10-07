package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
)

type Invite struct {
	UUID           uuid.UUID  `json:"uuid" validate:"uuid"`
	Email          string     `json:"email" validate:"email"`
	FederationUUID uuid.UUID  `json:"federation_uuid" validate:"uuid"`
	CompanyUUID    *uuid.UUID `json:"company_uuid" validate:"omitempty,uuid"`

	AcceptedAt *time.Time `json:"accepted_at"`
	DeclinedAt *time.Time `json:"declined_at"`

	CreatedAt time.Time  `json:"created_at"`
	UptatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func NewInvite(email string, federationUUID uuid.UUID, companyUUID *uuid.UUID) *Invite {
	invite := &Invite{
		UUID:           uuid.New(),
		Email:          email,
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		CreatedAt:      time.Now(),
		UptatedAt:      time.Now(),
	}

	errs, ok := helpers.ValidationStruct(invite)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return invite
}
