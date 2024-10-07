package app

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/agents"
	"github.com/krisch/crm-backend/internal/aggregates"
	"github.com/krisch/crm-backend/internal/cache"
	"github.com/krisch/crm-backend/internal/catalogs"
	"github.com/krisch/crm-backend/internal/comments"
	"github.com/krisch/crm-backend/internal/company"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/emails"
	"github.com/krisch/crm-backend/internal/federation"
	"github.com/krisch/crm-backend/internal/gates"
	"github.com/krisch/crm-backend/internal/health"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	"github.com/krisch/crm-backend/internal/logs"
	"github.com/krisch/crm-backend/internal/notifications"
	"github.com/krisch/crm-backend/internal/permissions"
	"github.com/krisch/crm-backend/internal/profile"
	"github.com/krisch/crm-backend/internal/reminders"
	"github.com/krisch/crm-backend/internal/s3"
	"github.com/krisch/crm-backend/internal/sms"
	"github.com/krisch/crm-backend/internal/task"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/sirupsen/logrus"
)

type App struct {
	Env  string
	Name string

	Port int

	Options configs.Configs

	LogService           logs.ILogService
	NotificationsService *notifications.Service
	ProfileService       *profile.Service
	EmailService         emails.IEmailsService
	HealthService        *health.Service
	FederationService    *federation.Service
	TaskService          *task.Service
	CommentService       *comments.Service
	DictionaryService    *dictionary.Service
	S3Service            *s3.Service
	S3PrivateService     *s3.ServicePrivate
	GateService          *gates.Service
	CacheService         *cache.Service
	RemindersService     *reminders.Service
	CatalogService       *catalogs.Service
	AgregateService      *aggregates.Service
	CompanyService       *company.Service
	SMSService           *sms.Service
	JWT                  jwt.IJWT
	AgentsService        *agents.Service
	PermissionsService   *permissions.Service

	MetricsCounters *helpers.MetricsCounters
}

func (a *App) SyncDictionariesByTimeout() {
	syncTime := time.Second * time.Duration(a.Options.DICTIONARY_SYNC_INTERVAL)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("exception: %s", string(debug.Stack()))
				time.Sleep(syncTime)
				a.SyncDictionariesByTimeout()
			}
		}()

		a.SyncDictionaries()
		a.DictionaryService.Info()
	}()
}

func (a *App) SyncDictionaries() {
	a.DictionaryService.SyncAll()
}

func (a *App) IsDev() bool {
	return a.Env == "dev" || a.Env == "test"
}

func (a *App) IsProd() bool {
	return a.Env == "prod"
}

func (a *App) SyncDictionariesByHook() {
	syncTime := time.Second

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("exception: %s", string(debug.Stack()))
				time.Sleep(syncTime)
				a.SyncDictionariesByHook()
			}
		}()

		for {
			if a.DictionaryService.ShoulUpdate() {
				a.DictionaryService.MarkToUpdate(false)
				a.SyncDictionaries()
				a.DictionaryService.Info()
			}

			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (a *App) RedisSubscribe(ctx context.Context, rds *redis.RDS, ch string) {
	pubsub := rds.Subscribe(ctx, ch)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("exception: %s", string(debug.Stack()))
				time.Sleep(time.Second * 5)
				a.RedisSubscribe(ctx, rds, ch)
			}
		}()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				logrus.Error(err)
			}

			logrus.WithField("redis", msg.Channel).
				Debug(msg.Payload)

			a.DictionaryService.MarkToUpdate(true)
		}
	}()
}

func (a *App) Work(ctx context.Context, rds *redis.RDS) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("exception: %s", string(debug.Stack()))
			time.Sleep(time.Second * 5)
			a.Work(ctx, rds)
		}
	}()

	a.RedisSubscribe(ctx, rds, "update")
	a.SyncDictionariesByTimeout()
	a.SyncDictionariesByHook()
}

func (a *App) Subscribe(_ context.Context) {
	a.TaskService.OnTaskUpdatedOrCreated(func(uid uuid.UUID, people []string) error {
		logrus.Info("task updated or created")
		err := a.NotificationsService.CreateTaskState(uid, people)
		return err
	})

	a.TaskService.OnOpenTask(func(uid uuid.UUID, email string) error {
		logrus.Info("task was open")
		err := a.NotificationsService.RemoveNotification(email, "task", uid)
		return err
	})

	a.RemindersService.OnReminderWasUpdatedOrCreated(func(uid, taskUUID uuid.UUID, people []string) error {
		logrus.Info("reminder updated or created: ", uid)
		err := a.NotificationsService.CreateTaskState(taskUUID, people)
		return err
	})
}
