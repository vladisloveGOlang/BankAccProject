//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/krisch/crm-backend/internal/activities"
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
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
)

func s3Conf(conf *configs.Configs) s3.Conf {
	return s3.Conf{
		Endpoint:        conf.CDN_PUBLIC_ENDPOINT,
		AccessKeyID:     conf.CDN_PUBLIC_ACCESS_KEY_ID,
		SecretAccessKey: conf.CDN_PUBLIC_SECRET_ACCESS_KEY,
		BucketName:      conf.CDN_PUBLIC_BUCKET_NAME,
		Location:        conf.CDN_PUBLIC_REGION,
		UseSSL:          conf.CDN_PUBLIC_SSL,
		PublicURL:       conf.CDN_PUBLIC_URL,
	}
}

func s3PrivateConf(conf *configs.Configs) s3.ConfPrivate {
	return s3.ConfPrivate{
		Endpoint:        conf.CDN_PRIVATE_ENDPOINT,
		AccessKeyID:     conf.CDN_PRIVATE_ACCESS_KEY_ID,
		SecretAccessKey: conf.CDN_PRIVATE_SECRET_ACCESS_KEY,
		BucketName:      conf.CDN_PRIVATE_BUCKET_NAME,
		Location:        conf.CDN_PRIVATE_REGION,
		UseSSL:          conf.CDN_PRIVATE_SSL,
		PublicURL:       conf.CDN_PRIVATE_URL,
		BackendURL:      conf.URL_BACKEND,
	}
}

func InitApp(name string, creds postgres.Creds, metrics bool, rc redis.Creds) (*App, error) {
	wire.Build(
		configs.NewConfigsFromEnv,
		postgres.NewGDB,
		redis.New,

		health.NewHealthService,

		dictionary.NewRepository,
		dictionary.New,

		federation.NewRepository,
		federation.NewUserService,

		cache.NewRepository,
		cache.New,

		notifications.NewRepository,
		notifications.New,

		activities.NewRepository,
		activities.New,

		logs.NewLogRepository,
		logs.NewLogService,

		company.NewRepository,
		company.New,

		permissions.NewRepository,
		permissions.New,

		s3Conf,
		s3.NewRepository,
		s3.New,

		agents.NewRepository,
		agents.New,

		s3PrivateConf,
		s3.NewPrivate,

		reminders.New,
		reminders.NewRepository,

		wire.Bind(new(gates.IDictionary), new(*dictionary.Service)),
		wire.Bind(new(dictionary.IStorage), new(*s3.Service)),

		sms.NewRepository,
		sms.New,

		gates.NewRepository,
		gates.New,

		emails.NewRepository,
		emails.NewFromCreds,

		helpers.NewMetricsCounters,

		aggregates.New,

		comments.NewRepository,
		comments.New,

		task.NewRepository,
		task.New,

		profile.NewRepository,
		profile.New,

		catalogs.NewRepository,
		catalogs.New,

		NewApp,
	)

	return &App{}, nil
}

func Configs() *configs.Configs {
	return &configs.Configs{}
}

func NewApp(name string, conf *configs.Configs, gdb *postgres.GDB, rds *redis.RDS,
	healthService *health.Service,
	notificationsService *notifications.Service,
	logService logs.ILogService,
	profileService *profile.Service,
	emailService emails.IEmailsService,
	federationService *federation.Service,
	taskService *task.Service,
	commentService *comments.Service,
	dictionaryService *dictionary.Service,
	s3Service *s3.Service,
	s3PrivateService *s3.ServicePrivate,
	gateService *gates.Service,
	cacheService *cache.Service,
	metricsCounters *helpers.MetricsCounters,
	remindersService *reminders.Service,
	catalogService *catalogs.Service,
	agregateService *aggregates.Service,
	companyService *company.Service,
	smsService *sms.Service,
	agentsService *agents.Service,
	permissionsService *permissions.Service,
) *App {
	w := &App{
		Env:  conf.ENV,
		Name: name,

		Port:    conf.PORT,
		Options: *conf,

		MetricsCounters: metricsCounters,
	}

	// Health
	w.HealthService = healthService
	w.NotificationsService = notificationsService
	w.LogService = logService
	w.EmailService = emailService

	// JWT service
	jwtService := jwt.New(conf.SOLT)
	jwtService.SetRefreshTokenValidator(func(token string) (bool, error) {
		return true, nil // @todo
	})
	jwtService.SetInvalidateToken(func(token string) (bool, error) {
		return true, nil // @todo
	})

	w.JWT = jwtService

	w.DictionaryService = dictionaryService
	w.GateService = gateService
	w.S3Service = s3Service
	w.S3PrivateService = s3PrivateService
	w.ProfileService = profileService
	w.CacheService = cacheService
	w.FederationService = federationService
	w.TaskService = taskService
	w.CommentService = commentService
	w.RemindersService = remindersService
	w.CatalogService = catalogService
	w.AgregateService = agregateService
	w.CompanyService = companyService
	w.SMSService = smsService
	w.AgentsService = agentsService
	w.PermissionsService = permissionsService

	return w
}
