package main

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/logs"
	"github.com/krisch/crm-backend/pkg/postgres"

	"github.com/sirupsen/logrus"
)

var (
	version   = "0.0.0"
	buildTime = "0.0.0"
	tag       = "-"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("exception: %s", string(debug.Stack()))
		}
	}()

	opt := configs.NewConfigsFromEnv()

	logs.InitLog(opt.LOG_FORMAT, opt.LOG_LEVEL, logs.SentryOptions{
		Enable:   opt.SENTRY_ENABLE,
		DSN:      opt.SENTRY_DSN,
		LogLevel: opt.LOG_LEVEL,
	})

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
}
