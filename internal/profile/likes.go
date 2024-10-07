package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type LikeEntityType string

const (
	Project    LikeEntityType = "project"
	Company    LikeEntityType = "company"
	Federation LikeEntityType = "federation"
	Task       LikeEntityType = "task"
	Redminder  LikeEntityType = "reminder"
)

func (s *Service) Like(userUUID uuid.UUID, entityType LikeEntityType, entityUUID uuid.UUID) error {
	// @todo: trunaction?
	key := fmt.Sprintf("user:%s:likes:%s", userUUID, entityType)
	err := s.repo.rds.SAddString(context.Background(), key, entityUUID.String(), 300000)
	if err != nil {
		return err
	}

	key2 := fmt.Sprintf("%s:%s:liked", entityType, entityUUID.String())
	err = s.repo.rds.SAddString(context.Background(), key2, userUUID.String(), 300000)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Dislike(userUUID uuid.UUID, entityType LikeEntityType, entityUUID uuid.UUID) error {
	// @todo: trunaction?
	key := fmt.Sprintf("user:%s:likes:%s", userUUID, entityType)
	err := s.repo.rds.SRemString(context.Background(), key, entityUUID.String())
	if err != nil {
		return err
	}

	key2 := fmt.Sprintf("%s:%s:liked", entityType, entityUUID.String())

	err = s.repo.rds.SRemString(context.Background(), key2, userUUID.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetUserLike(userUUID uuid.UUID, entityType string) (uids []uuid.UUID, err error) {
	uids = make([]uuid.UUID, 0)

	key := fmt.Sprintf("user:%s:likes:%s", userUUID, entityType)
	likes, err := s.repo.rds.SGet(context.Background(), key)
	if err != nil {
		return uids, err
	}

	for _, like := range likes {
		uid, err := uuid.Parse(like)
		if err == nil {
			uids = append(uids, uid)
		} else {
			logrus.Warnf("Failed to parse like %s key: %s", like, key)
		}
	}

	return uids, err
}

func (s *Service) CheckEntityHasLike(entityType string, entityUUID, userUUID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("%s:%s:liked", entityType, entityUUID.String())
	isLiked, err := s.repo.rds.SIsMember(context.Background(), key, userUUID.String())
	if err != nil {
		return false, err
	}

	return isLiked, err
}
