package logs

import (
	"encoding/json"
	"time"
)

type Log struct {
	UUID             string    `gorm:"type:uuid;default:gen_random_uuid()"`
	BackendUUID      string    `gorm:"type:string;default:unknown"`
	HeaderXRequestID string    `gorm:"type:varchar(255);default:''"`
	Message          string    `gorm:"type:text"`
	IP               string    `gorm:"type:varchar(255)"`
	Host             string    `gorm:"type:varchar(255)"`
	Method           string    `gorm:"type:varchar(255)"`
	RequestURI       string    `gorm:"type:varchar(255)"`
	Status           int       `gorm:"type:integer"`
	Request          string    `gorm:"type:jsonb;default:'{}'"`
	Agent            string    `gorm:"type:varchar(255)"`
	Referer          string    `gorm:"type:varchar(255)"`
	Start            time.Time `gorm:"type:timestamp with time zone"`
	Stop             time.Time `gorm:"type:timestamp with time zone"`
	Token            string    `gorm:"type:text"`
}

func (l *Log) ToJSON() string {
	e, err := json.Marshal(l)
	if err != nil {
		return ""
	}

	return string(e)
}
