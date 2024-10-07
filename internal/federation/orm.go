package federation

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/profile"
	"gorm.io/datatypes"
)

type Federation struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;unique:true"`
	ID             int64     `gorm:"primary_key:true"`
	Name           string    `gorm:"type:varchar(50);default:'';not null"`
	CreatedBy      string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByBUUID uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	FederationUsers []FederationUser `gorm:"foreignKey:FederationUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Companies       []Company        `gorm:"foreignKey:FederationUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Company struct {
	UUID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name           string     `gorm:"type:varchar(100);default:'';not null"`
	FederationUUID uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedBy      string     `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID  uuid.UUID  `gorm:"type:uuid;not null;"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt      *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	FieldLastName int `gorm:"type:int;default:0;not null"`

	CompamyUsers []CompanyUser `gorm:"foreignKey:CompanyUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Federation   Federation    `gorm:"foreignKey:FederationUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	UsersTotal    int `gorm:"type:int;default:0;->"`
	ProjectsTotal int `gorm:"type:int;default:0;->"`
}

type Project struct {
	UUID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name           string     `gorm:"type:varchar(100);default:'';not null"`
	Description    string     `gorm:"type:text;default:'';not null"`
	FederationUUID uuid.UUID  `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedBy      string     `gorm:"type:varchar(100);default:'';not null;"`
	ResponsibleBy  string     `gorm:"type:varchar(100);default:'';not null;"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt      *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	StatusGraph string                `gorm:"type:jsonb;default:'{}';not null"`
	Options     domain.ProjectOptions `gorm:"type:jsonb;default:'{}';not null"`

	Status          int        `gorm:"type:int;default:0;not null;"`
	Stops           Stops      `gorm:"type:jsonb;default:'[]';not null;"`
	StatusUpdatedAt *time.Time `gorm:"->;type:timestamptz;"`

	StatusSort IntArray    `gorm:"type:jsonb;default:'[]';not null;column:status_sort"`
	FieldsSort StringArray `gorm:"type:jsonb;default:'[]';not null;column:fields_sort"`
}

type ProjectStatistic struct {
	UUID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();"`

	TasksTotal         int `gorm:"type:int;default:0;->"`
	TasksFinishedTotal int `gorm:"type:int;default:0;->"`
	TasksDeletedTotal  int `gorm:"type:int;default:0;->"`
	TasksCanceledTotal int `gorm:"type:int;default:0;->"`
	TasksActiveTotal   int `gorm:"type:int;default:0;->"`
}

type CompanyFields struct {
	UUID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Hash        string    `gorm:"type:varchar(15);not null;"`
	Name        string    `gorm:"type:varchar(100);not null;"`
	Description string    `gorm:"type:text;not null;"`
	Icon        string    `gorm:"type:varchar(50);not null;"`
	DataType    int       `gorm:"type:int;not null;default:0"`
	CompanyUUID uuid.UUID `gorm:"type:uuid;not null"`

	ProjectUUID JSONArray `gorm:"->;type:jsonb;default:'[]';not null;column:project_uuids"`

	TasksTotal        int `gorm:"type:int;default:0;->"`
	TasksFilled       int `gorm:"type:int;default:0;->"`
	TasksActiveTotal  int `gorm:"type:int;default:0;->"`
	TasksActiveFilled int `gorm:"type:int;default:0;->"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	RequiredOnStatuses IntArray `gorm:"->;type:jsonb;default:'[]';not null;"`
	Style              string   `gorm:"->;type:varchar(20);default:'';not null;"`
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

type StringArray []string

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []string{}
	err := json.Unmarshal(bytes, &result)
	*j = StringArray(result)
	return err
}

// Value return json value, implement driver.Valuer interface.
func (j StringArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

//nolint:revive //it is orm table name
type FederationUser struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_member"`
	UserUUID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_member"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`

	Federation Federation   `gorm:"foreignKey:FederationUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User       profile.User `gorm:"foreignKey:UserUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CompanyUser struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null;"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null;"`
	UserUUID       uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`

	Federation Federation   `gorm:"foreignKey:FederationUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Company    Company      `gorm:"foreignKey:CompanyUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User       profile.User `gorm:"foreignKey:UserUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ProjectUser struct {
	UUID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	ProjectUUID uuid.UUID `gorm:"type:uuid;not null;"`
	UserUUID    uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`

	Project Project      `gorm:"foreignKey:ProjectUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User    profile.User `gorm:"foreignKey:UserUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type User struct {
	UUID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name    string    `gorm:"type:varchar(50);default:'';not null"`
	Lname   string    `gorm:"type:varchar(50);default:'';not null"`
	Pname   string    `gorm:"type:varchar(50);default:'';not null"`
	Email   string    `gorm:"type:varchar(100);default:'';not null;unique:true"`
	Phone   int       `gorm:"type:int;default:0;not null;"`
	IsValid bool      `gorm:"type:boolean;default:false;not null"`

	Color    string `gorm:"type:varchar(7);default:'#000000';not null"`
	HasPhoto bool   `gorm:"type:boolean;default:false;not null"`

	AddedToFederation *time.Time `gorm:"type:timestamptz;default:NULL;"`
	AddedToCompany    *time.Time `gorm:"type:timestamptz;default:NULL;"`
	AddedToProject    *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type Invite struct {
	UUID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Email          string     `gorm:"type:varchar(100);default:'';not null"`
	FederationUUID uuid.UUID  `gorm:"type:uuid;not null;"`
	CompanyUUID    *uuid.UUID `gorm:"type:uuid;null;"`

	AcceptedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
	DeclinedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type ProjectCatalogData struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null;"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null;"`
	ProjectUUID    uuid.UUID `gorm:"type:uuid;not null;"`

	Name  string `gorm:"type:varchar(100);default:'';not null"`
	Value string `gorm:"type:varchar(500);default:'';not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type Group struct {
	UUID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name string    `gorm:"type:varchar(100);default:'';not null"`

	FederationUUID uuid.UUID `gorm:"type:uuid;not null"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	UserUuids StringArray `gorm:"->"`
}

type GroupUser struct {
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	GroupUUID uuid.UUID `gorm:"type:uuid;not null;"`
	UserUUID  uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type ProjectField struct {
	UUID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	CompanyUUID      uuid.UUID `gorm:"type:uuid;not null;"`
	ProjectUUID      uuid.UUID `gorm:"type:uuid;not null;"`
	CompanyFieldUUID uuid.UUID `gorm:"type:uuid;not null;"`

	RequiredOnStatuses IntArray `gorm:"type:jsonb;default:'[]';not null;"`
	Style              string   `gorm:"type:varchar(50);default:'';not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type ProjectStatus struct {
	UUID        *uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name        string     `gorm:"type:varchar(100);default:'';not null"`
	Number      int        `gorm:"type:int;default:10;not null"`
	Color       string     `gorm:"type:text;default:'';not null"`
	Description string     `gorm:"type:text;default:'';not null"`
	CompanyUUID uuid.UUID  `gorm:"type:uuid;not null"`
	ProjectUUID uuid.UUID  `gorm:"type:uuid;not null"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type Stop struct {
	CreatedAt     time.Time `json:"created_at"`
	StatusID      int       `json:"status_id"`
	CreatedByUUID uuid.UUID `json:"created_by_uuid"`
}

type Stops []Stop

func (j *Stops) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []Stop{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j Stops) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type FieldStatistics struct {
	Hash   string  `json:"hash"`
	Name   string  `json:"name"`
	Count  int     `json:"count"`
	Total  int     `json:"total"`
	Filled float64 `json:"filled"`
}
