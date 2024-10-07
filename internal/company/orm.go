package company

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

//nolint:revive // it is table name
type CompanyTags struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name           string    `gorm:"type:varchar(100);default:'';not null"`
	Color          string    `gorm:"type:text;default:'';not null"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

//nolint:revive // it is table name
type CompanyPriority struct {
	UUID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name        string    `gorm:"type:varchar(100);default:'';not null"`
	Number      int       `gorm:"type:int;default:10;not null"`
	Color       string    `gorm:"type:text;default:'';not null"`
	CompanyUUID uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type Company struct {
	UUID       uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	SmsOptions SmsOptions `gorm:"type:jsonb;default:'';not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type SmsOptions struct {
	API  string `json:"api"`
	From string `json:"from"`
}

func (j *SmsOptions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := SmsOptions{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j SmsOptions) Value() (driver.Value, error) {
	return json.Marshal(j)
}
