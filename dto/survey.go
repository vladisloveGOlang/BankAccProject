package dto

import (
	"time"

	"github.com/google/uuid"
)

type SurveyDTO struct {
	UUID uuid.UUID `json:"uuid"`

	User UserDTO `json:"user"`

	Name string                 `json:"name"`
	Body map[string]interface{} `json:"body"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type SurveyDTOs struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
