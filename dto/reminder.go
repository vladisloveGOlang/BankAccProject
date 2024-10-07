package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReminderDTO struct {
	UUID        uuid.UUID  `json:"uuid"`
	TaskUUID    uuid.UUID  `json:"task_uuid"`
	Comment     string     `json:"comment"`
	Description string     `json:"description"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	Type        string     `json:"type"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Status int `json:"status"`

	User      *UserDTO `json:"user,omitempty"`
	CreatedBy *UserDTO `json:"created_by,omitempty"`
}
