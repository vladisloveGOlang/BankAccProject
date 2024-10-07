package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/aggregates"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/redis"
	v9 "github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

type Repository struct {
	rds *redis.RDS
}

func NewRepository(rds *redis.RDS) *Repository {
	return &Repository{
		rds: rds,
	}
}

// Task.
func (r *Repository) StoreTaskState(email, kindWithUUID string, state aggregates.StateDiff) error {
	key := fmt.Sprintf("notifications:%s:%s", email, kindWithUUID)

	js, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = r.rds.HSET(context.Background(), key, "state", js)
	if err != nil {
		return err
	}

	//

	lastOpen, err := r.rds.HGet(context.Background(), key, "last_open")
	if err != nil {
		return err
	}

	if lastOpen == "" {
		err = r.rds.HSET(context.Background(), key, "last_open", time.Now().UnixMicro())
		if err != nil {
			return err
		}
	}

	//

	key = fmt.Sprintf("notifications:%s", email)
	score := state.UpdatedAt.UnixMicro()
	err = r.rds.ZADD(context.Background(), key, kindWithUUID, score)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ToggleStarNotification(email, kindWithUUID string, star bool) error {
	key := fmt.Sprintf("notifications:%s:%s", email, kindWithUUID)

	err := r.rds.HSET(context.Background(), key, "star", helpers.If(star, "1", "0"))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) HideNotification(email, kindWithUUID string) error {
	key := fmt.Sprintf("notifications:%s:%s", email, kindWithUUID)

	err := r.rds.HSET(context.Background(), key, "last_open", time.Now().UnixMicro())
	if err != nil {
		return err
	}

	err = r.RemoveNotification(email, kindWithUUID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetTaskStateNotification(email, kindWithUUID string) (map[string]string, error) {
	key := fmt.Sprintf("notifications:%s:%s", email, kindWithUUID)
	return r.rds.HGetAll(context.Background(), key)
}

func (r *Repository) StoreNotification(email, kind string, uid uuid.UUID) error {
	key := fmt.Sprintf("notifications:%s", email)

	tm := time.Now().UnixMicro()

	return r.rds.ZADD(context.Background(), key, kind+":"+uid.String(), tm)
}

func (r *Repository) IncNotification(email, kind, counter string, uid uuid.UUID) error {
	key := fmt.Sprintf("notifications:%s:count:%s", email, kind+":"+uid.String())

	return r.rds.HIncrBy(context.Background(), key, counter)
}

func (r *Repository) IncNotificationCount(email, kind string, uid uuid.UUID) error {
	key := fmt.Sprintf("notifications:%s", email)

	tm := time.Now().UnixMicro()

	return r.rds.ZADD(context.Background(), key, kind+":"+uid.String(), tm)
}

func (r *Repository) RemoveNotification(email, kindWithUUID string) error {
	key := fmt.Sprintf("notifications:%s", email)

	err := r.rds.ZREM(context.Background(), key, kindWithUUID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetLastNotificationTime(email, uuidWithKind string) (t time.Time, found bool, err error) {
	key := fmt.Sprintf("notifications:%s:%s", email, uuidWithKind)

	lastOpen, err := r.rds.HGet(context.Background(), key, "last_open")
	if errors.Is(err, v9.Nil) {
		return time.UnixMicro(0), false, nil
	}

	if err != nil {
		return time.UnixMicro(0), false, err
	}

	if lastOpen == "" {
		return time.UnixMicro(0), false, nil
	}

	lastOpenInt, err := strconv.ParseInt(lastOpen, 10, 64)
	if err != nil {
		return time.UnixMicro(0), false, err
	}

	return time.UnixMicro(lastOpenInt), true, nil
}

func (r *Repository) GetNotification(email string) (dtos []dto.NotificationDTO, err error) {
	key := fmt.Sprintf("notifications:%s", email)

	z, err := r.rds.ZRange(context.Background(), key, 120)
	if err != nil {
		return dtos, err
	}

	dtos = lo.Map(z, func(v v9.Z, i int) dto.NotificationDTO {
		uuidWithKind, ok := v.Member.(string)
		if !ok {
			return dto.NotificationDTO{}
		}

		star, err := r.rds.HGet(context.Background(), key+":"+uuidWithKind, "star")
		if err != nil {
			return dto.NotificationDTO{}
		}

		if !strings.Contains(uuidWithKind, ":") {
			return dto.NotificationDTO{
				UUID:  uuidWithKind,
				Type:  "-",
				Score: v.Score,
				Star:  star == "1",
			}
		}

		kind := strings.Split(uuidWithKind, ":")[0]
		uid := strings.Split(uuidWithKind, ":")[1]

		return dto.NotificationDTO{
			UUID:  uid,
			Type:  kind,
			Score: v.Score,
			Star:  star == "1",
		}
	})

	return dtos, nil
}

func (r *Repository) GetNotificationCount(email, kind string, uid uuid.UUID) (map[string]string, error) {
	key := fmt.Sprintf("notifications:%s:count:%s", email, kind+":"+uid.String())

	return r.rds.HGetAll(context.Background(), key)
}

func (r *Repository) RemoveNotifications(ctx context.Context, email string) error {
	key := fmt.Sprintf("notifications:%s", email)

	return r.rds.Del(ctx, key)
}

func (r *Repository) Count(email string) (v int64, err error) {
	key := fmt.Sprintf("notifications:%s", email)

	v, err = r.rds.ZCount(context.Background(), key)

	return v, err
}
