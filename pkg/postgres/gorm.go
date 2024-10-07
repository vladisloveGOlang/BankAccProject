package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/krisch/crm-backend/internal/helpers"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"gorm.io/plugin/prometheus"
)

type GDB struct {
	DB *gorm.DB
}

func NewGDB(creds Creds, metrics bool) (*GDB, error) {
	connStr, err := helpers.ConvertPostgresCreds(string(creds))
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		TranslateError:  true,
		CreateBatchSize: 1000,
		Logger:          NewLogger(),
	})
	if err != nil {
		return nil, err
	}

	if metrics {
		err = db.Use(prometheus.New(prometheus.Config{
			DBName:          "db-dev",
			RefreshInterval: 15,
			MetricsCollector: []prometheus.MetricsCollector{
				&prometheus.Postgres{
					VariableNames: []string{"Threads_running"},
				},
			},
		}))

		if err != nil {
			return nil, err
		}
	}

	return &GDB{DB: db}, nil
}

type GormLogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	Debug                 bool
}

func NewLogger() *GormLogger {
	return &GormLogger{
		SkipErrRecordNotFound: true,
		Debug:                 true,
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	logrus.WithContext(ctx).Infof(s, args...)
}

func (l *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	logrus.WithContext(ctx).Warnf(s, args...)
}

func (l *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	logrus.WithContext(ctx).Errorf(s, args...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[logrus.ErrorKey] = err
		logrus.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		logrus.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}

	if l.Debug {
		logrus.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
	}
}
