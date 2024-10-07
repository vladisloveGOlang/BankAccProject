package domain

import (
	"time"

	"github.com/google/uuid"
)

type Survey struct {
	UUID uuid.UUID

	User User

	Name string
	Body map[string]interface{}

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
