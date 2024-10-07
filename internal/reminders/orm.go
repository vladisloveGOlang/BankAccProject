package reminders

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	UUID          uuid.UUID  `gorm:"<-:create;type:uuid;primary_key"`
	Description   string     `gorm:"type:text"`
	Comment       string     `gorm:"type:text"`
	CreatedBy     string     `gorm:"<-:create;type:varchar(255)"`
	CreatedByUUID uuid.UUID  `gorm:"<-:create;type:varchar(255)"`
	UserUUID      *uuid.UUID `gorm:"type:uuid"`
	TaskUUID      uuid.UUID  `gorm:"<-:create;type:uuid"`
	DateFrom      *time.Time `gorm:"type:timestamp"`
	DateTo        *time.Time `gorm:"type:timestamp"`
	Type          string     `gorm:"type:varchar(50)"`
	Status        int        `gorm:"type:integer"`

	CreatedAt time.Time `gorm:"->;type:timestamp"`
	UpdatedAt time.Time
	DeletedAt *time.Time
}
