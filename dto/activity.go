package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/sirupsen/logrus"
)

type ActivityDTO struct {
	UUID      uuid.UUID               `json:"uuid"`
	CreatedBy UserDTO                 `json:"created_by,omitempty"`
	CreatedAt time.Time               `json:"created_at"`
	Meta      *map[string]interface{} `json:"meta,omitempty"`
	Type      int                     `json:"type"`

	TaskStatus *ActivityTaskStatusDTO `json:"task_status,omitempty"`
	TaskField  *ActivityTaskFieldDTO  `json:"task_field,omitempty"`
	Status     map[string]interface{} `json:"status,omitempty"`
}

type ActivityTaskStatusDTO struct {
	Old       int               `json:"old"`
	New       int               `json:"new"`
	OldStatus ProjectStatusDTOs `json:"old_status"`
	NewStatus ProjectStatusDTOs `json:"new_status"`
}

type ActivityTaskFieldDTO struct {
	Old  interface{} `json:"old"`
	New  interface{} `json:"new"`
	Name string      `json:"name"`
}

type ActivityTaskFieldArrayDTO struct {
	Old  []interface{} `json:"old"`
	New  []interface{} `json:"new"`
	Name string        `json:"name"`

	Add    []interface{} `json:"add"`
	Remove []interface{} `json:"remove"`
}

type ActivityTaskTeamDTO struct {
	Add    []UserDTO `json:"add"`
	Remove []UserDTO `json:"remove"`
	Name   string    `json:"name"`
}

type ActivityTaskWasDeletedDTO struct {
	Name string `json:"name"`
}

type ActivityTaskFileWasDeletedDTO struct {
	Name string `json:"name"`
	Ext  string `json:"ext"`
	Size int64  `json:"size"`
}

func NewActivityDTO(dm domain.Activity, user UserDTO) *ActivityDTO {
	var status map[string]interface{}

	if dm.Type == int(domain.ActivityTaskStatus) {
		var p ActivityTaskStatusDTO
		metaBytes, err := json.Marshal(dm.Meta)
		if err != nil {
			logrus.Error("cannot marshal meta")
		} else {
			err = json.Unmarshal(metaBytes, &p)
			if err != nil {
				logrus.Error("cannot unmarshal meta")
			} else {
				status, err = helpers.StructToMap(&p)
				if err != nil {
					logrus.Error("cannot convert struct to map")
				}
			}
		}
	}

	if dm.Type == int(domain.ActivityTaskField) {
		var p ActivityTaskFieldDTO
		metaBytes, err := json.Marshal(dm.Meta)
		if err != nil {
			logrus.Error("cannot marshal meta")
		} else {
			err = json.Unmarshal(metaBytes, &p)
			if err != nil {
				logrus.Error("cannot unmarshal meta")
			} else {
				status, err = helpers.StructToMap(&p)
				if err != nil {
					logrus.Error("cannot convert struct to map")
				}
			}
		}
	}

	if dm.Type == int(domain.ActivityTaskTeamArray) {
		var p ActivityTaskTeamDTO
		metaBytes, err := json.Marshal(dm.Meta)
		if err != nil {
			logrus.Error("cannot marshal meta")
		} else {
			err = json.Unmarshal(metaBytes, &p)
			if err != nil {
				logrus.Error("cannot unmarshal meta")
			} else {
				status, err = helpers.StructToMap(&p)
				if err != nil {
					logrus.Error("cannot convert struct to map")
				}
			}
		}
	}

	if dm.Type == int(domain.ActivityTaskWasDeleted) {
		var p ActivityTaskWasDeletedDTO
		metaBytes, err := json.Marshal(dm.Meta)
		if err != nil {
			logrus.Error("cannot marshal meta")
		} else {
			err = json.Unmarshal(metaBytes, &p)
			if err != nil {
				logrus.Error("cannot unmarshal meta")
			} else {
				status, err = helpers.StructToMap(&p)
				if err != nil {
					logrus.Error("cannot convert struct to map")
				}
			}
		}
	}

	if dm.Type == int(domain.ActivityTaskFileWasDeleted) {
		var p ActivityTaskFileWasDeletedDTO
		metaBytes, err := json.Marshal(dm.Meta)
		if err != nil {
			logrus.Error("cannot marshal meta")
		} else {
			err = json.Unmarshal(metaBytes, &p)
			if err != nil {
				logrus.Error("cannot unmarshal meta")
			} else {
				status, err = helpers.StructToMap(&p)
				if err != nil {
					logrus.Error("cannot convert struct to map")
				}
			}
		}
	}

	return &ActivityDTO{
		UUID:      dm.UUID,
		CreatedBy: user,
		CreatedAt: dm.CreatedAt,
		Type:      dm.Type,

		Status: status,
	}
}
