package dto

import (
	"time"

	"github.com/google/uuid"
)

type TagDTO struct {
	UUID uuid.UUID `json:"uuid"`

	Name  string `json:"name"`
	Color string `json:"color"`

	CreatedBy UserDTO `json:"created_by,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type TagDTOs struct {
	UUID uuid.UUID `json:"uuid"`

	Name  string `json:"name"`
	Color string `json:"color"`
}
