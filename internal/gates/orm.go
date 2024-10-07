package gates

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
)

type Permission struct {
	UserUUID       uuid.UUID `json:"user_uuid" gorm:"<-:create;"`
	FederationUUID uuid.UUID `json:"federation_uuid" gorm:"<-:create;"`

	Rules domain.PermissionRules `json:"rules"`

	CreatedAt time.Time  `json:"created_at" gorm:"<-:create;"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type JSON domain.PermissionRules

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := domain.PermissionRules{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}
