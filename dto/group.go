package dto

import (
	"time"

	"github.com/google/uuid"
)

type GroupDTO struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserUUIDS []string `json:"user_uuids"`
}

type GroupDTOs struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}
