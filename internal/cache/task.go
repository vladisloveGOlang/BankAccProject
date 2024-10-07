package cache

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/sirupsen/logrus"
)

func (s *Service) CacheTask(ctx context.Context, task *dto.TaskDTO) {
	if s.cacheTaskTTL == 0 {
		logrus.WithField("cache", "task").Debug("cache is disabled")
		return
	}

	uid := task.UUID

	taskJSONToStore, err := json.Marshal(&task)
	if err != nil {
		logrus.Error("GetTask: ", err)
	}

	err = s.repo.rds.SetStr(ctx, uid.String(), string(taskJSONToStore), s.cacheTaskTTL)
	if err != nil {
		logrus.Error("GetTask: ", err)
	}

	logrus.Debugf("cache stored for %vs", s.cacheTaskTTL)
}

func (s *Service) GetTask(ctx context.Context, uid uuid.UUID) (task dto.TaskDTO, err error) {
	if s.cacheTaskTTL == 0 {
		return task, errors.New("cache is disabled")
	}

	taskJSON, err := s.repo.rds.GetStr(ctx, uid.String())

	if err == nil && taskJSON != "" {
		err = json.Unmarshal([]byte(taskJSON), &task)
		if err == nil {
			logrus.Info("[module:cache] GetTask: from redis")
			return task, nil
		}
		logrus.Error("[module:cache] GetTask: unmarshal err: ", err)
	} else if err != nil {
		logrus.Error("[module:cache] GetTask: ", err)
	}

	return task, nil
}

func (s *Service) ClearTask(ctx context.Context, uid uuid.UUID) {
	logrus.Debugf("[module:cache] clear task cache for uuid: %v", uid)
	err := s.repo.rds.Del(ctx, uid.String())
	if err != nil {
		logrus.Error("[module:cache] ClearTask: ", err)
	}
}
