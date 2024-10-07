package dictionary

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID `gorm:"type:uuid;default:'';not null"`
	Name      string    `gorm:"type:varchar(50);default:'';not null"`
	Lname     string    `gorm:"type:varchar(50);default:'';not null"`
	Pname     string    `gorm:"type:varchar(50);default:'';not null"`
	Email     string    `gorm:"type:varchar(100);default:'';not null;unique:true"`
	Phone     int       `gorm:"type:int;default:0;not null;"`
	UpdatedAt time.Time `gorm:"type:timestamptz;"`
	HasPhoto  bool      `gorm:"type:boolean;default:false;not null"`

	FederationUUID uuid.UUID `gorm:"->;type:federation_uuid"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type UserFederation struct {
	UserUUID       uuid.UUID `gorm:"type:user_uuid;not null"`
	FederationUUID uuid.UUID `gorm:"type:federation_uuid;not null"`
	CreatedAt      time.Time `gorm:"type:timestamptz;"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type UsersCompanies struct {
	UserUUID    uuid.UUID `gorm:"type:user_uuid;not null"`
	CompanyUUID uuid.UUID `gorm:"type:company_uuid;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type CompanyTags struct {
	UUID  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name  string    `gorm:"type:varchar(100);default:'';not null"`
	Color string    `gorm:"type:text;default:'';not null"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	UpdatedAt time.Time  `gorm:"type:timestamptz;"`
	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type UsersProjects struct {
	UserUUID    uuid.UUID `gorm:"type:user_uuid;not null"`
	ProjectUUID uuid.UUID `gorm:"type:project_uuid;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type Federation struct {
	UUID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;unique:true"`
	ID            int64     `gorm:"primary_key:true"`
	Name          string    `gorm:"type:varchar(50);default:'';not null"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;"`
	CreatedByUUID uuid.UUID `gorm:"type:created_by_b_uuid;not null"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type Company struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name           string    `gorm:"type:varchar(100);default:'';not null"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type Project struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name           string    `gorm:"type:varchar(100);default:'';not null"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;"`

	StatusGraph string `gorm:"type:jsonb;not null"`
	Options     string `gorm:"type:jsonb;not null"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type CompanyFields struct {
	Hash        string    `gorm:"type:varchar(15);not null;"`
	Name        string    `gorm:"type:varchar(100);not null;"`
	DataType    int       `gorm:"type:int;not null;default:0"`
	CompanyUUID uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;"`

	DeletedAt *time.Time `gorm:"type:timestamptz;"`

	ProjectsUUID UUIDArray `gorm:"type:jsonb;default:'[]';not null;"`
}

type ProjectFields struct {
	Hash        string    `gorm:"type:varchar(15);not null;"`
	Name        string    `gorm:"type:varchar(100);not null;"`
	DataType    int       `gorm:"type:int;not null;default:0"`
	CompanyUUID uuid.UUID `gorm:"type:uuid;not null"`
	ProjectUUID uuid.UUID `gorm:"type:uuid;not null"`

	UpdatedAt time.Time  `gorm:"type:timestamptz;"`
	DeletedAt *time.Time `gorm:"type:timestamptz;"`

	RequiredOnStatuses IntArray `gorm:"type:jsonb;default:'[]';not null;"`
}

type UUIDArray []uuid.UUID

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *UUIDArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []uuid.UUID{}
	err := json.Unmarshal(bytes, &result)
	*j = UUIDArray(result)
	return err
}

// Value return json value, implement driver.Valuer interface.
func (j UUIDArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type CatalogFields struct {
	Hash        string    `gorm:"type:varchar(15);not null;"`
	Name        string    `gorm:"type:varchar(100);not null;"`
	DataType    int       `gorm:"type:int;not null;default:0"`
	CatalogUUID uuid.UUID `gorm:"type:uuid;not null"`

	UpdatedAt time.Time  `gorm:"type:timestamptz;"`
	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type CompanyPriority struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Number      int       `json:"priority"`
	Color       string    `json:"color"`
	CompanyUUID uuid.UUID `json:"company_uuid"`

	UpdatedAt time.Time  `gorm:"type:timestamptz;"`
	DeletedAt *time.Time `gorm:"type:timestamptz;"`
}

type IntArray []int

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *IntArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []int{}
	err := json.Unmarshal(bytes, &result)
	*j = IntArray(result)
	return err
}

// Value return json value, implement driver.Valuer interface.
func (j IntArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}
