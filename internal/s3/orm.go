package s3

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	UUID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key:true"`

	Type     string    `gorm:"type:int;not null"`
	TypeUUID uuid.UUID `gorm:"type:uuid;not null"`

	Name       string `gorm:"type:varchar(50);default:'';not null"`
	ObjectName string `gorm:"type:varchar(250);default:'';not null"`
	Size       int64  `gorm:"type:bigint;default:0;not null"`

	ImgResized bool `gorm:"type:boolean;default:false;not null"`
	ImgWidth   int  `gorm:"type:int;default:0;not null"`
	ImgHeight  int  `gorm:"type:int;default:0;not null"`

	Ext        string `gorm:"type:varchar(10);default:'';not null"`
	MimeType   string `gorm:"type:varchar(20);default:'';not null"`
	BucketName string `gorm:"type:varchar(200);default:'';not null"`
	Endpoint   string `gorm:"type:varchar(30);default:'';not null"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null;"`

	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt   *time.Time `gorm:"type:timestamptz;default:NULL;"`
	ToDeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`
}
