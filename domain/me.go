package domain

import (
	"github.com/google/uuid"
)

type Me struct {
	UUID  uuid.UUID
	Email string `validate:"email"`
}
