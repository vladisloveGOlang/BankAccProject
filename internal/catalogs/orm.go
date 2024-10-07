package catalogs

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type JSONArray []any

func (j JSONArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type Catalog struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;unique:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null"`
	Name           string    `validate:"lte=100,gte=3"  ru:"название"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	FieldLastName int `gorm:"type:int;default:0;not null"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`
}

type CatalogFields struct {
	UUID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Hash            string     `gorm:"type:varchar(15);not null;"`
	Name            string     `gorm:"type:varchar(100);not null;"`
	DataType        int        `gorm:"type:int;not null;default:0"`
	DataCatalogUUID *uuid.UUID `gorm:"type:uuid"`
	CatalogUUID     uuid.UUID  `gorm:"type:uuid;not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type CatalogData struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null"`
	CatalogUUID    uuid.UUID `gorm:"type:uuid;not null"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	Fields JSONB `gorm:"default:'{}';not null;"`

	Entities JSONArray `gorm:"type:jsonb;default:'{}';not null;"`

	EntitiesRich JSONArray `gorm:"<-:false;type:jsonb;default:'{}';not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null" order:""`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null" order:""`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Total int64 `gorm:"->"`
}
