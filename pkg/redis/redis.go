package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RDS struct {
	rdb *redis.Client
}

type Creds string

func New(creds Creds) (*RDS, error) {
	// parse string: redis://$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable

	pattern := regexp.MustCompile(`redis://(?P<password>[^@]+)@(?P<host>[^:]+):(?P<port>[^/]+)/(?P<dbname>[^?]+)`)
	sub := pattern.FindStringSubmatch(string(creds))

	if len(sub) != 5 {
		return nil, fmt.Errorf("invalid redis connection string")
	}

	password := sub[1]
	host := sub[2]

	port, err := strconv.Atoi(sub[3])
	if err != nil {
		return nil, err
	}

	dbIndex, err := strconv.Atoi(sub[4])
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%v", host, port),
		Password: password,
		DB:       dbIndex,
	})

	return &RDS{rdb: rdb}, nil
}

// Fetch str from redis.
func (rds *RDS) GetStr(ctx context.Context, key string) (string, error) {
	v, err := rds.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return v, nil
}

// ttl - in seconds.
func (rds *RDS) SetStr(ctx context.Context, key, value string, ttl int) error {
	err := rds.rdb.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()

	return err
}

func (rds *RDS) Del(ctx context.Context, key string) error {
	err := rds.rdb.Del(ctx, key).Err()

	return err
}

func (rds *RDS) SAddString(ctx context.Context, key, value string, ttl int) error {
	err := rds.rdb.SAdd(ctx, key, value).Err()
	if err != nil {
		return err
	}

	err = rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()

	return err
}

func (rds *RDS) Expire(ctx context.Context, key string, ttl int) error {
	return rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
}

func (rds *RDS) SGet(ctx context.Context, key string) ([]string, error) {
	return rds.rdb.SMembers(ctx, key).Result()
}

func (rds *RDS) SRemString(ctx context.Context, key, value string) error {
	err := rds.rdb.SRem(ctx, key, value).Err()

	return err
}

func (rds *RDS) SCARD(ctx context.Context, key string) (int64, error) {
	return rds.rdb.SCard(ctx, key).Result()
}

func (rds *RDS) SInterString(ctx context.Context, col1, col2 string) ([]string, error) {
	res, err := rds.rdb.SInter(ctx, col1, col2).Result()
	if err != nil {
		return []string{}, err
	}

	return res, nil
}

func (rds *RDS) SIsMember(ctx context.Context, col1, col2 string) (bool, error) {
	f, err := rds.rdb.SIsMember(ctx, col1, col2).Result()
	if err != nil {
		return f, err
	}

	return f, nil
}

func (rds *RDS) ZADD(ctx context.Context, key, memberUUID string, score int64) error {
	z := redis.Z{
		Score:  float64(score),
		Member: memberUUID,
	}
	return rds.rdb.ZAdd(ctx, key, z).Err()
}

func (rds *RDS) ZREM(ctx context.Context, key, value string) error {
	return rds.rdb.ZRem(ctx, key, value).Err()
}

func (rds *RDS) ZCount(ctx context.Context, key string) (v int64, err error) {
	v, err = rds.rdb.ZCount(ctx, key, "-inf", "+inf").Result()
	if err != nil {
		return v, err
	}

	return v, err
}

func (rds *RDS) ZRange(ctx context.Context, key string, last int64) (v []redis.Z, err error) {
	v, err = rds.rdb.ZRangeWithScores(ctx, key, last*-1, -1).Result()

	return v, err
}

func (rds *RDS) HSET(ctx context.Context, key string, name, value interface{}) (err error) {
	err = rds.rdb.HSet(ctx, key, name, value).Err()

	return err
}

func (rds *RDS) HIncrBy(ctx context.Context, key, name string) (err error) {
	err = rds.rdb.HIncrBy(ctx, key, name, int64(1)).Err()

	return err
}

func (rds *RDS) HDel(ctx context.Context, key string) error {
	return rds.rdb.HDel(ctx, key).Err()
}

func (rds *RDS) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rds.rdb.HGetAll(ctx, key).Result()
}

func (rds *RDS) HGet(ctx context.Context, key, field string) (string, error) {
	res, err := rds.rdb.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}

	return res, err
}

func (rds *RDS) Publish(ctx context.Context, channel, message string) error {
	return rds.rdb.Publish(ctx, channel, message).Err()
}

func (rds *RDS) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return rds.rdb.Subscribe(ctx, channel)
}

func (rds *RDS) WConnetcion(ctx context.Context) error {
	err := rds.rdb.Set(ctx, "last_connection", time.Now().Format(time.RFC3339), time.Hour*24*365).Err()

	return err
}

func (rds *RDS) Ping(ctx context.Context) error {
	err := rds.rdb.Ping(ctx).Err()

	return err
}
