//nolint
package sms

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type sms struct {
	UUID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null:false;unique:true"`

	FederationUUID uuid.UUID `gorm:"type:uuid;not null;"`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;"`
	CreatedByUUID uuid.UUID `gorm:"type:uuid;not null;"`

	To     string  `gorm:"type:varchar(100);default:'';not null;"`
	From   string  `gorm:"type:varchar(100);default:'';not null;"`
	Text   string  `gorm:"type:text;default:'';not null;"`
	Status string  `gorm:"type:varchar(100);default:'';not null;"`
	Cost   float64 `gorm:"type:float;default:0;not null;"`

	CreatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	Total int64 `gorm:"->"`
}
