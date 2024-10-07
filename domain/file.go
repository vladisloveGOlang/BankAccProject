package domain

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	Name string    `json:"name"`
	Ext  string    `json:"ext"`
	Size int64     `json:"int"`
	UUID uuid.UUID `json:"uuid"`
	URL  string    `json:"url"`

	CreatedAt time.Time `json:"created_at"`
	CreatedBy uuid.UUID `json:"created_by"`
}
