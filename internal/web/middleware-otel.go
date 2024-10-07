package web

import (
	"fmt"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	//nolint // jaeger is the only supported exporter at the moment.
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	// ScopeName is the instrumentation scope name.
	ID = 6666
)

func tracerProvider(url, service, environment string) *trace.TracerProvider {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil
	}
	tp := trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exp),
		// Record information about this application in a Resource.
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", ID),
		)),
	)
	return tp
}

// Middleware returns echo middleware which will trace incoming requests.

func TraceMiddleware(service, url, env string, opts ...Option) echo.MiddlewareFunc {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = tracerProvider(url, service, env)
	}
	tracer := cfg.TracerProvider.Tracer(
		service,
	)

	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			c.Set("tracer", tracer)

			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				request = request.WithContext(savedCtx)
				c.SetRequest(request)
			}()
			ctx := cfg.Propagators.Extract(savedCtx, propagation.HeaderCarrier(request.Header))

			spanName := c.Path()
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s route not found", request.Method)
			}

			ctx, span := tracer.Start(ctx, spanName)
			defer span.End()

			c.SetRequest(request.WithContext(ctx))

			err := next(c)
			if err != nil {
				span.SetAttributes(attribute.String("echo.error", err.Error()))
				c.Error(err)
			}

			status := c.Response().Status
			span.SetStatus(http.StatusContinue, "")
			if status > 0 {
				span.SetAttributes(semconv.HTTPStatusCode(status))
			}

			return err
		}
	}
}

// config is used to configure the mux middleware.
type config struct {
	TracerProvider *trace.TracerProvider
	Propagators    propagation.TextMapPropagator
	Skipper        middleware.Skipper
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *config) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

// WithSkipper specifies a skipper for allowing requests to skip generating spans.
func WithSkipper(skipper middleware.Skipper) Option {
	return optionFunc(func(cfg *config) {
		cfg.Skipper = skipper
	})
}
