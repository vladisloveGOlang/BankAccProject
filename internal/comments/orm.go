package comments

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key:true"`
	Comment   string    `gorm:"type:varchar(500);default:'';not null"`
	CreatedBy string    `gorm:"type:varchar(100);default:'';not null;"`

	TaskUUID uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null;"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now();not null;"`

	ReplyUUID    *uuid.UUID `gorm:"type:uuid;"`
	ReplyComment *string    `gorm:"->;type:varchar(500);omitempty"`

	Likes  Persons `gorm:"type:jsonb;default:'{}';not null;"`
	People Persons `gorm:"type:text[];default:'{}';not null;"`

	Pin bool `gorm:"type:boolean;default:false;not null;"`
}

// JSONB Interface for JSONB Field of yourTableName Table.
type Persons map[string]int64

// Value Marshal.
func (a Persons) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan Unmarshal.
func (a *Persons) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
