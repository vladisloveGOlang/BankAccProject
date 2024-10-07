package main

import (
	"context"
	"os"
	"runtime/debug"
	"time"

	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/helpers"
	logs "github.com/krisch/crm-backend/internal/logs"
	"github.com/krisch/crm-backend/internal/web"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"

	"github.com/sirupsen/logrus"
)

var (
	version   = "0.0.0"
	tag       = "-"
	buildTime = "0.0.0"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("exception: %s", string(debug.Stack()))
		}
	}()

	ctx := context.Background()

	opt := configs.NewConfigsFromEnv()

	logs.InitLog(opt.LOG_FORMAT, opt.LOG_LEVEL, logs.SentryOptions{
		Enable:   opt.SENTRY_ENABLE,
		DSN:      opt.SENTRY_DSN,
		LogLevel: opt.LOG_LEVEL,
	})

	setTimzone(opt.TIME_ZONE)

	go func() {
		for {
			logrus.Warn("start mem usage")
			helpers.PrintMemUsage()

			time.Sleep(time.Second * 60)
		}
	}()

	logrus.Debug("------------------")
	logrus.
		WithField("version", version).
		WithField("tag", tag).
		WithField("buildTime", buildTime).
		WithField("now", helpers.DateNow()).
		Infof("starting server v%v (%v) from %v", version, tag, buildTime)

	opt.Debug()

	//

	if _, err := os.Stat("/tmp"); os.IsNotExist(err) {
		err := os.Mkdir("/tmp", os.ModePerm)
		logrus.Error(err)
	}

	//

	if opt.MIGRATE {
		logrus.Debug("migrating...")
		pdb, err := postgres.NewPDB(opt.DB_CREDS)
		if err != nil {
			logrus.Error(err)
		}

		err = pdb.Migrate(opt.MIGRATE_FOLDER)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Info("migrating done")
		}

		time.Sleep(time.Second * 5)
	}

	//

	w := web.NewWeb(*opt)
	w.Version = version
	w.Tag = tag
	w.BuildTime = buildTime

	rds, err := redis.New(opt.REDIS_CREDS)
	if err != nil {
		logrus.Error(err)
	}

	go sendAliveToRedis(w.UUID, rds)

	go w.Work(ctx, rds)

	w.Init()
	w.Run()
}

func sendAliveToRedis(name string, rds *redis.RDS) {
	ctx := context.Background()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("exception: %s", string(debug.Stack()))
				time.Sleep(time.Second * 15)
				sendAliveToRedis(name, rds)
			}
		}()

		for {
			time.Sleep(time.Second * 5)

			err := rds.SetStr(ctx, helpers.ToSnake("last_connaction:"+name), time.Now().Format(time.RFC3339), 5)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
	}()
}

func setTimzone(timeZone string) {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		logrus.Error(err)
	}

	time.Local = loc
}
