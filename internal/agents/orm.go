package agents

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Agent struct {
	UUID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;unique:true"`

	FederationUUID uuid.UUID  `gorm:"type:uuid;not null;"`
	CompanyUUID    *uuid.UUID `gorm:"type:uuid;not null;"`
	ProjectUUID    *uuid.UUID `gorm:"type:uuid;"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	Name     string        `gorm:"type:varchar(100);default:'';not null;"`
	Contacts ContactsArray `gorm:"type:jsonb;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	Total int64 `gorm:"->"`
}

type Contacts struct {
	Type string `json:"type"`
	Val  string `json:"val"`
}

type ContactsArray []Contacts

func (j *ContactsArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []Contacts{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j ContactsArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}
