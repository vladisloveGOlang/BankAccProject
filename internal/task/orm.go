package task

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Task struct {
	UUID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key:true"`
	ID            int       `gorm:"primaryKey" order:""`
	Name          string    `gorm:"type:varchar(50);default:'';not null" order:""`
	CreatedBy     string    `gorm:"type:varchar(100);default:'';not null;" order:""`
	ResponsibleBy string    `gorm:"type:varchar(100);default:'';not null;" order:""`
	ImplementBy   string    `gorm:"type:varchar(100);default:'';not null;" order:""`
	ManagedBy     string    `gorm:"type:varchar(100);default:'';not null;" order:""`
	FinishedBy    string    `gorm:"type:varchar(100);default:'';not null;" order:""`

	CoWorkersBy pq.StringArray `gorm:"type:text[];default:'{}';not null;"`
	WatchBy     pq.StringArray `gorm:"type:text[];default:'{}';not null;"`
	AllPeople   pq.StringArray `gorm:"type:text[];default:'{}';not null;"`

	Tags pq.StringArray `gorm:"type:text[];default:'{}';not null;"`

	FederationUUID uuid.UUID `gorm:"type:uuid;not null" order:""`
	CompanyUUID    uuid.UUID `gorm:"type:uuid;not null" order:""`
	ProjectUUID    uuid.UUID `gorm:"type:uuid;not null" order:""`

	Icon     string `gorm:"type:varchar(20);default:'';not null"`
	Status   int    `gorm:"type:int;default:0;not null" order:""`
	Priority int    `gorm:"type:int;default:10;not null" order:""`

	IsEpic         bool           `gorm:"type:bool;default:false;not null;" order:""`
	ChildrensTotal int            `gorm:"type:int;default:0;not null;" order:""`
	ChildrensUUID  pq.StringArray `gorm:"type:uuid[];default:'{}';not null;"`

	CommentsTotal int `gorm:"type:int;default:0;not null;" order:""`

	CreatedAt  time.Time  `gorm:"type:timestamptz;default:now();not null" order:""`
	FinishedAt *time.Time `gorm:"type:timestamptz;default:NULL;" order:""`
	FinishTo   *time.Time `gorm:"type:timestamptz;default:NULL;" order:""`
	ActivityAt time.Time  `gorm:"type:timestamptz;default:now();not null" order:""`

	Duration int `gorm:"type:int;default:0;not null"`

	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now();not null" order:""`
	DeletedAt *time.Time `gorm:"type:timestamptz;default:NULL;"`

	Fields JSONB `gorm:"type:jsonb;default:'{}';not null;"`

	Meta datatypes.JSON `gorm:"default:'{}';not null;"`

	Path string `gorm:"type:ltree;default:'';"`

	Total int64 `gorm:"->"`

	TaskEntities TE    `gorm:"type:jsonb;default:'{}';not null;"`
	Stops        Stops `gorm:"type:jsonb;default:'[]';not null;"`

	FirstOpen FirstOpen `gorm:"->update;type:jsonb;default:'{}';not null;"`

	Description string `gorm:"type:text;default:'';not null" order:""`
}

type FirstOpen map[string]time.Time

func (j FirstOpen) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *FirstOpen) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type TE map[uuid.UUID][]string

func (j *TE) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := TE{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j TE) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Stop struct {
	UUID          uuid.UUID
	CreatedAt     time.Time
	StatusID      int
	StatusName    string
	Comment       string
	CreatedBy     string
	CreatedByUUID uuid.UUID
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

type Project struct {
	UUID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key:true"`
	TaskID int       `gorm:"type:int;not null;"`
}

type CompanyFields struct {
	Hash        string `gorm:"type:varchar(15);not null;"`
	Name        string `gorm:"type:varchar(100);not null;"`
	DataType    int    `gorm:"type:int;not null;default:0"`
	CompanyUUID string `gorm:"type:uuid;not null"`
}
