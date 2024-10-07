module github.com/krisch/crm-backend

go 1.21.5

toolchain go1.21.7

require (
	github.com/brianvoe/gofakeit/v6 v6.26.4
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/disintegration/gift v1.2.1
	github.com/gabriel-vasile/mimetype v1.4.3
	github.com/go-playground/assert/v2 v2.2.0
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/gorilla/websocket v1.5.1
	github.com/iancoleman/strcase v0.3.0
	github.com/jellydator/ttlcache/v3 v3.1.1
	github.com/minio/minio-go/v7 v7.0.66
	github.com/oapi-codegen/runtime v1.1.1
	github.com/samber/lo v1.39.0
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/otel/sdk v1.26.0
	go.opentelemetry.io/otel/trace v1.26.0
	golang.org/x/crypto v0.22.0
	golang.org/x/sync v0.7.0
	golang.org/x/time v0.5.0
	gorm.io/gorm v1.25.9
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20210507211836-431795d63e8d // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.2 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/microsoft/go-mssqldb v1.6.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.46.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xuri/efp v0.0.0-20230802181842-ad255f2331ca // indirect
	github.com/xuri/nfp v0.0.0-20230819163627-dc951e3ffe1a // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/mysql v1.5.2 // indirect
)

require (
	github.com/caarlos0/env/v6 v6.10.1
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/go-playground/validator/v10 v10.16.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.6.0
	github.com/google/wire v0.5.0
	github.com/labstack/echo-contrib v0.15.0
	github.com/labstack/echo/v4 v4.11.4
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.18.0
	github.com/redis/go-redis/v9 v9.4.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/xuri/excelize/v2 v2.8.0
	go.opentelemetry.io/otel v1.26.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.uber.org/atomic v1.11.0
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gorm.io/datatypes v1.2.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/plugin/prometheus v0.1.0
)

replace github.com/Shopify/sarama => github.com/IBM/sarama v1.43.1
