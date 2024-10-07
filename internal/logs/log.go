package logs

import (
	"database/sql"
	"strings"
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"
)

type LogService struct {
	repo ILogRepository
}

func NewLogService(repo ILogRepository) ILogService {
	return &LogService{
		repo: repo,
	}
}

type ILogService interface {
	InsertLog(l Log) error
}

// ----------------------------

func (s *LogService) InsertLog(l Log) error {
	return s.repo.InsertLog(l)
}

// ----------------------------

type SentryOptions struct {
	Enable   bool
	DSN      string
	LogLevel string
}

func InitLog(format, level string, sentryOptions SentryOptions) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	h := dbHook{}
	logrus.AddHook(h)

	//

	initSentry(sentryOptions)
}

type Hook struct{}

// Add a hook to an instance of logger. This is called with
// `log.Hooks.Add(new(MyHook))` where `MyHook` implements the `Hook` interface.
//nolint
func (hooks Hook) Add(hook Hook) {
}

// Fire all the hooks for the passed level. Used by `entry.log` to fire
// appropriate hooks for a log entry.
//nolint
func (hooks Hook) Fire(level logrus.Level, entry *logrus.Entry) error {
	return nil
}

//

type dbHook struct {
	*sql.DB
}

func (db dbHook) Fire(e *logrus.Entry) error {
	if e.Message == "" {
		return nil
	}

	x, y, z := 0, 0, 0
	res := ""
	for i, c := range e.Message[:len(e.Message)-1] {
		if c == '[' {
			x = i
		}

		if c == ']' {
			z = i
			y = i

			tag := e.Message[x+1 : y]

			if strings.Contains(tag, ":") {
				tagWithValue := strings.Split(tag, ":")
				e.Data[tagWithValue[0]] = tagWithValue[1]
			} else {
				e.Data["tag"] = tag
			}
		}
	}

	if z != 0 {
		res += strings.Trim(e.Message[z+1:], " ")
	}

	if res == "" {
		res = e.Message
	}

	e.Message = res

	return nil
}

func (dbHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func initSentry(opt SentryOptions) {
	if !opt.Enable {
		logrus.Debug("sentry disabled")
		return
	}

	if opt.DSN == "" {
		logrus.Error("sentry dsn is wrong")
		return
	}

	if opt.LogLevel == "info" {
		hook, err := logrus_sentry.NewSentryHook(opt.DSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.InfoLevel,
		})
		if err != nil {
			logrus.Error(err)
		}
		logrus.AddHook(hook)
	} else {
		hook, err := logrus_sentry.NewSentryHook(opt.DSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})
		if err != nil {
			logrus.Error(err)
		}

		hook.Timeout = 10 * time.Second
		hook.StacktraceConfiguration.Enable = true

		logrus.AddHook(hook)
	}
}
