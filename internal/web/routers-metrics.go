package web

import (
	"log"

	"github.com/labstack/echo-contrib/echoprometheus"
	echo "github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func initMetricsRoutes(a *Web, e *echo.Echo) {
	e.Use(echoprometheus.NewMiddleware(a.Options.APP_NAME))

	customCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
	)

	if err := prometheus.Register(customCounter); err != nil {
		log.Fatal(err)
	}

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		AfterNext: func(c echo.Context, err error) {
			customCounter.Inc()
		},
	}))

	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
}
