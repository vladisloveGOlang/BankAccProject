package notifications

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/aggregates"
	"github.com/sirupsen/logrus"
)

func (s *Service) CreateTaskState(uid uuid.UUID, people []string) error {
	task, err := s.aggs.GetTaskWithFields(context.TODO(), uid)
	if err != nil {
		logrus.Error("send task updated or created error: ", err)
		return err
	}

	for _, p := range people {
		user, ok := s.dict.FindUser(p)
		if !ok {
			logrus.Errorf("user not found: %s", p)
			continue
		}

		if task.DeletedAt != nil {
			err := s.repo.RemoveNotification(p, "task:"+uid.String())
			if err != nil {
				logrus.Error("RemoveNotification error: ", err)
			}
			continue
		}

		t, found, err := s.repo.GetLastNotificationTime(p, "task:"+uid.String())
		if err != nil {
			logrus.Error("GetLastNotificationTime error: ", err)
			continue
		}

		if !found {
			t = time.Unix(0, 0)
		}

		diffState := aggregates.CompareState(task, user.UUID, t)

		err = s.repo.StoreTaskState(p, "task:"+uid.String(), diffState)
		if err != nil {
			logrus.Error("StoreTaskState error: ", err)
		}
	}

	return nil
}

func (s *Service) GetTaskState(email string, uid uuid.UUID) (state aggregates.StateDiff, star bool, err error) {
	js, err := s.repo.GetTaskStateNotification(email, "task:"+uid.String())
	if err != nil {
		return state, star, err
	}

	if js["star"] == "1" {
		star = true
	}

	err = json.Unmarshal([]byte(js["state"]), &state)
	if err != nil {
		return state, star, err
	}

	return state, star, nil
}
