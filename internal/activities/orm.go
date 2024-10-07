package activities

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
)

type Activity struct {
	UUID uuid.UUID `gorm:"type:uuid;default:'';not null"`

	EntityUUID uuid.UUID `gorm:"type:uuid;default:'';not null"`
	EntityType string    `gorm:"type:varchar(255);default:'';not null"`

	CreatedByUUID uuid.UUID `gorm:"type:uuid;default:'';not null"`
	CreatedBy     string    `gorm:"type:varchar(255);default:'';not null"`

	Meta Meta `gorm:"type:jsonb;default:'{}'::jsonb;not null"`

	Description string    `gorm:"type:varchar(255);default:'';not null"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:now();not null"`

	Type domain.ActivityType `gorm:"type:integer;default:0;not null"`

	Total int64 `gorm:"->"`
}

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
