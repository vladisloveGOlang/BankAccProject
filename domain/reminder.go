package domain

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	UUID          uuid.UUID
	Description   string     `validate:"lte=5000"  ru:"описание"`
	Comment       string     `validate:"lte=5000"  ru:"комментарий"`
	CreatedBy     string     `validate:"lte=100,gte=3"  ru:"автор (email)"`
	CreatedByUUID uuid.UUID  `validate:"uuid"  ru:"автор (uuid)"`
	TaskUUID      uuid.UUID  `validate:"uuid"  ru:"задача (uuid)"`
	DateFrom      *time.Time `ru:"дата"`
	DateTo        *time.Time `ru:"дата"`
	Type          string     `validate:"gte=1,lte=50"  ru:"тип"`
	UserUUID      *uuid.UUID `validate:"uuid"  ru:"пользователь (uuid)"`
	Status        int        `validate:"gte=0,lte=10"  ru:"статус"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
