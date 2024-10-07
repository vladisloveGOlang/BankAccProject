package emails

import (
	"time"

	"gorm.io/datatypes"
)

type Mail struct {
	UUID    string `gorm:"type:uuid;default:gen_random_uuid();not null:false;primary_key:true"`
	From    string `gorm:"type:varchar(100);default:'';not null"`
	To      string `gorm:"type:varchar(100);default:'';not null;"`
	Subject string `gorm:"type:varchar(200);default:'';not null;"`
	Text    string `gorm:"type:text;default:'';not null;"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`
}
