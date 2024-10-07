package cache

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (s *Service) CacheURL(ctx context.Context, fileUUID uuid.UUID, url string) {
	if s.cachePresignedURLsTTL == 0 {
		logrus.WithField("cache", "presigned_url").Debug("cache is disabled")
		return
	}

	err := s.repo.rds.SetStr(ctx, fileUUID.String(), url, s.cachePresignedURLsTTL)
	if err != nil {
		logrus.Error("GetTask: ", err)
	}

	logrus.Debugf("cache stored for %vs", s.cachePresignedURLsTTL)
}

func (s *Service) GetURL(ctx context.Context, fileUUID uuid.UUID) (url string, err error) {
	if s.cachePresignedURLsTTL == 0 {
		return "", errors.New("cache is disabled")
	}

	url, err = s.repo.rds.GetStr(ctx, fileUUID.String())

	return url, err
}

func (s *Service) ClearURL(ctx context.Context, fileUUID uuid.UUID) {
	logrus.Debugf("[module:cache] clear url cache for uuid: %v", fileUUID)
	err := s.repo.rds.Del(ctx, fileUUID.String())
	if err != nil {
		logrus.Error("[module:cache] ClearURL: ", err)
	}
}
