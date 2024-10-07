package profile

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"gorm.io/datatypes"
)

type User struct {
	UUID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Name     string    `gorm:"type:varchar(50);default:'';not null"`
	Lname    string    `gorm:"type:varchar(50);default:'';not null"`
	Pname    string    `gorm:"type:varchar(50);default:'';not null"`
	Email    string    `gorm:"type:varchar(100);default:'';not null;unique:true"`
	Phone    int       `gorm:"type:int;default:0;not null;"`
	IsValid  bool      `gorm:"type:boolean;default:false;not null"`
	Password string    `gorm:"type:varchar(100);default:'';not null"`
	Provider int       `gorm:"type:int;default:0;not null"`
	Color    string    `gorm:"type:varchar(7);default:'#000000';not null"`
	HasPhoto bool      `gorm:"type:boolean;default:false;not null"`

	CreatedAt        time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt        *time.Time `gorm:"type:timestamptz;default:NULL;"`
	ValidAt          *time.Time `gorm:"type:timestamptz;default:NULL;"`
	ValidationSendAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Preferences UserPreferences `gorm:"default:'{}';not null;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`
}

type UserPreferences struct {
	Timezone *string `json:"timezone,omitempty"`
}

func (j *UserPreferences) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := UserPreferences{
		Timezone: helpers.Ptr("Europe/Moscow"),
	}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j UserPreferences) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Invite struct {
	UUID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	Email          string     `gorm:"type:varchar(100);default:'';not null;unique:true"`
	FederationUUID uuid.UUID  `gorm:"type:uuid;default:NULL;"`
	CompanyUUID    *uuid.UUID `gorm:"type:uuid;default:NULL;"`

	AcceptedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
	DeclinedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

type FederationUser struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_member"`
	UserUUID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_member"`
}

type CompanyUser struct {
	UUID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	FederationUUID uuid.UUID `gorm:"type:uuid;not null;"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null;"`
	UserUUID       uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
}

type Preference struct {
	UUID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	UserUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user"`

	Preferences datatypes.JSON `gorm:"default:'{}';not null;"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
}

type Survey struct {
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	UserUUID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user"`
	UserEmail string    `gorm:"type:varchar(100);default:'';not null"`
	Name      string    `gorm:"type:varchar(50);default:'';not null"`
	Body      JSONB     `gorm:"default:'{}';not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}

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
