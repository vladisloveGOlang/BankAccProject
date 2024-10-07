package configs

import (
	"fmt"
	"reflect"

	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"

	env "github.com/caarlos0/env/v6"
)

func NewConfigsFromEnv() *Configs {
	o := Configs{}

	return o.loadFromEnv()
}

//nolint // uniq code style
type Configs struct {
	// ENV - dev, prod, test
	ENV string `env:"ENV" default:"prod"`

	// DB
	DB_CREDS postgres.Creds `env:"DB_CREDS" secured:"true"`

	// Redis
	REDIS_CREDS redis.Creds `env:"REDIS_CREDS" secured:"true"`

	// SMTP
	SMTP_ENABLE bool   `env:"SMTP_ENABLE" envDefault:"true"`
	SMTP_CREDS  string `env:"SMTP_CREDS" secured:"true"`

	// APP
	GZIP                     int    `env:"GZIP" envDefault:"5"`
	LOG_LEVEL                string `env:"LOG_LEVEL" envDefault:"debug"`
	LOG_FORMAT               string `env:"LOG_FORMAT" envDefault:"plain"`
	LOG_TO_DB                bool   `env:"LOG_TO_DB" envDefault:"true"`
	APP_NAME                 string `env:"APP_NAME" envDefault:"unknown"`
	SOLT                     string `env:"SOLT" envDefault:"solt"`
	TIME_ZONE                string `env:"TIME_ZONE" envDefault:"UTC"`
	DICTIONARY_SYNC_INTERVAL int    `env:"DICTIONARY_SYNC_INTERVAL" envDefault:"10"`
	URL_BACKEND              string `env:"URL_BACKEND" envDefault:"http://localhost:8080"`

	// CDN
	CDN_PUBLIC_REGION            string `env:"CDN_PUBLIC_REGION" envDefault:"us-east-1"`
	CDN_PUBLIC_ENDPOINT          string `env:"CDN_PUBLIC_ENDPOINT" envDefault:"storage.yandexcloud.net"`
	CDN_PUBLIC_ACCESS_KEY_ID     string `env:"CDN_PUBLIC_ACCESS_KEY_ID" envDefault:""`
	CDN_PUBLIC_SECRET_ACCESS_KEY string `env:"CDN_PUBLIC_SECRET_ACCESS_KEY" envDefault:"" secured:"true"`
	CDN_PUBLIC_BUCKET_NAME       string `env:"CDN_PUBLIC_BUCKET_NAME" envDefault:""`
	CDN_PUBLIC_SSL               bool   `env:"CDN_PUBLIC_SSL" envDefault:"true"`
	CDN_PUBLIC_URL               string `env:"CDN_PUBLIC_URL" envDefault:"https://storage.yandexcloud.net"`

	CDN_PRIVATE_REGION            string `env:"CDN_PRIVATE_REGION" envDefault:"us-east-1"`
	CDN_PRIVATE_ENDPOINT          string `env:"CDN_PRIVATE_ENDPOINT" envDefault:"storage.yandexcloud.net"`
	CDN_PRIVATE_ACCESS_KEY_ID     string `env:"CDN_PRIVATE_ACCESS_KEY_ID" envDefault:""`
	CDN_PRIVATE_SECRET_ACCESS_KEY string `env:"CDN_PRIVATE_SECRET_ACCESS_KEY" envDefault:"" secured:"true"`
	CDN_PRIVATE_BUCKET_NAME       string `env:"CDN_PRIVATE_BUCKET_NAME" envDefault:""`
	CDN_PRIVATE_SSL               bool   `env:"CDN_PRIVATE_SSL" envDefault:"true"`
	CDN_PRIVATE_URL               string `env:"CDN_PRIVATE_URL" envDefault:"https://storage.yandexcloud.net"`

	// Features
	SEED           bool   `env:"SEED" envDefault:"false"`
	METRICS        bool   `env:"METRICS" envDefault:"true"`
	MIGRATE        bool   `env:"MIGRATE" envDefault:"false"`
	MIGRATE_FOLDER string `env:"MIGRATE_FOLDER" envDefault:"./migrations"`
	RATE_LIMITER   int    `env:"RATE_LIMITER" envDefault:"20"`

	// Sentry
	SENTRY_DSN    string `env:"SENTRY_DSN" secured:"true"`
	SENTRY_ENABLE bool   `env:"SENTRY_ENABLE" envDefault:"false"`

	// Telemetry
	OTEL_ENABLE   bool   `env:"OTEL_ENABLE" envDefault:"false"`
	OTEL_EXPORTER string `env:"OTEL_EXPORTER" envDefault:"http://localhost:14268/api/traces"`

	// Cache
	CACHE_TASK            int `env:"CACHE_TASKS" envDefault:"15"`
	CACHE_TASKS           int `env:"CACHE_TASKS" envDefault:"15"`
	CACHE_PRE_SIGNED_URLS int `env:"CACHE_PRE_SIGNED_URLS" envDefault:"15"`

	// HTTP
	PORT int `env:"PORT" envDefault:"8080"`

	CORS_ENABLE            bool   `env:"CORS_ENABLE" envDefault:"false"`
	CORS_ALLOW_CREDENTIALS bool   `env:"CORS_ALLOW_CREDENTIALS" envDefault:"false"`
	CORS_ALLOWED_ORIGINS   string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	// SMS
	SMS_API_ID string `env:"SMS_API_ID" secured:"true"`
	SMS_FROM   string `env:"SMS_FROM" envDefault:"sector"`

	// Integration
	MAX_EMAIL_MONTHS           int      `env:"MAX_EMAIL_MONTHS" envDefault:"1"`
	EMAILS_INTEGRATION_ENABLED bool     `env:"EMAILS_INTEGRATION_ENABLED" envDefault:"false"`
	KAFKA_BROKERS              []string `env:"KAFKA_BROKERS" envDefault:"kafka:9092"`
	KAFKA_TOPIC                string   `env:"KAFKA_TOPIC" envDefault:"emails"`
}

func (o *Configs) Debug() {
	for key, value := range o.GetFieldsWithValues() {
		logrus.Debug(key + ": " + value)
	}
}

func (o *Configs) GetSecuredFilds() []string {
	secureFields := []string{}

	e := reflect.ValueOf(o).Elem()

	for i := 0; i < e.NumField(); i++ {
		field := e.Type().Field(i)

		if getStructTag(field, "secured") != "" {
			secureFields = append(secureFields, field.Name)
		}
	}

	return secureFields
}

func (o *Configs) GetFieldsWithValues() map[string]string {
	result := make(map[string]string)

	secureFields := o.GetSecuredFilds()

	e := reflect.ValueOf(o).Elem()
	for i := 0; i < e.NumField(); i++ {
		field := e.Type().Field(i)

		if helpers.InArray(field.Name, secureFields) {
			value := e.Field(i).String()

			switch lvalue := len(value); {
			case lvalue == 0:
				result[field.Name] = ""
			case lvalue > 5:
				value = value[:len(value)/10*7]
				result[field.Name] = value + "..."
			case lvalue > 0:
				result[field.Name] = "xxxxxxx"
			}
		} else {
			result[field.Name] = fmt.Sprintf("%v", e.Field(i).Interface())
		}
	}

	return result
}

func getStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}

func (o *Configs) loadFromEnv() *Configs {
	options := Configs{}
	if err := env.Parse(&options); err != nil {
		panic(err)
	}

	return &options
}

func (o *Configs) NewKafkaWriter() *kafka.Writer {
	if len(o.KAFKA_BROKERS) == 0 {
		panic("Kafka brokers not configured")
	}
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: o.KAFKA_BROKERS,
		Topic:   o.KAFKA_TOPIC,
	})
}
