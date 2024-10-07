package permissions

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Group struct {
	UUID  uuid.UUID `json:"uuid" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name  string    `json:"name" gorm:"unique;not null"`
	State Meta      `json:"permissions" gorm:"type:jsonb;default:'{}'::jsonb;not null"`

	CreatedBy uuid.UUID `json:"created_by"`

	CreatedAt time.Time  `json:"created_at" gorm:"<-:create;"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (u *Group) TableName() string {
	return "permissions.groups"
}

type User struct {
	UserUUID uuid.UUID `json:"user_uuid"`

	State  Meta      `json:"state"`
	Groups uuid.UUID `json:"groups"`

	CreatedAt time.Time  `json:"created_at" gorm:"<-:create;"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (u *User) TableName() string {
	return "permissions.users"
}

type Meta map[string]interface{}

func (j *Meta) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ActivityMeta value:", value))
	}

	result := Meta{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j Meta) Value() (driver.Value, error) {
	return json.Marshal(j)
}
