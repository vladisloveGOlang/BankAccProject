package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/app"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/redis"

	validator "github.com/go-playground/validator/v10"
)

type Web struct {
	app     *app.App
	Options configs.Configs
	Router  *echo.Echo
	Port    int

	UUID string

	Now       string
	Version   string
	Tag       string
	BuildTime string
}

func NewWeb(conf configs.Configs) *Web {
	name := helpers.FakeName()

	a, err := app.InitApp(name, conf.DB_CREDS, true, conf.REDIS_CREDS)
	if err != nil {
		logrus.Fatal(err)
	}

	return &Web{
		app:     a,
		Options: conf,
		Now:     helpers.DateNow(),
		UUID:    name,

		Port: conf.PORT,
	}
}

func (a *Web) Work(ctx context.Context, rds *redis.RDS) {
	a.app.Work(ctx, rds)
	a.app.Subscribe(ctx)
}

var upgrader = websocket.Upgrader{}

func hello(a *Web, _ *echo.Echo) func(c echo.Context) error {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		for {
			// Read
			_, msg, err := ws.ReadMessage()
			if err != nil {
				logrus.Error(err)
				continue
			}

			fmt.Printf("%s\n", msg)

			arr := strings.Split(string(msg), " ")

			if len(arr) < 2 {
				continue
			}

			search := domain.SearchUser{
				FederationUUID: uuid.MustParse(arr[0]),
				Search:         arr[1],
			}

			dmns, err := a.app.FederationService.SearchUserInDictionary(search)
			if err != nil {
				logrus.Error(err)
				continue
			}

			dtos := lo.Map(dmns, func(item domain.User, index int) dto.UserDTO {
				return dto.NewUserDto(item, a.app.ProfileService)
			})

			jsn, err := json.Marshal(dtos)
			if err != nil {
				logrus.Error(err)
				continue
			}

			err = ws.WriteMessage(websocket.TextMessage, jsn)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
	}
}

func (a *Web) Init() *echo.Echo {
	e := echo.New()

	// Middlewares
	if a.Options.CORS_ENABLE {
		origins := strings.Split(a.Options.CORS_ALLOWED_ORIGINS, ",")

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     origins,
			AllowCredentials: a.Options.CORS_ALLOW_CREDENTIALS,
			AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))
	}

	if a.Options.OTEL_ENABLE {
		e.Use(TraceMiddleware("crm", a.Options.OTEL_EXPORTER, a.Options.ENV, WithSkipper(middleware.DefaultSkipper)))
	}

	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("150M"))
	e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
		Skipper: func(c echo.Context) bool {
			if strings.Contains(c.Request().RequestURI, "/comment") {
				return true
			}

			if strings.Contains(c.Request().RequestURI, "/profile/photo") {
				return true
			}

			return false
		},
		Limit: "2M",
	}))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.BodyDump(LogMiddleware(a.app)))

	if a.Options.GZIP > 0 {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: a.app.Options.GZIP,
			Skipper: func(c echo.Context) bool {
				return c.Request().RequestURI == "/metrics"
			},
		}))
	}

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var notFoundErr dto.NotFoundError
		if errors.As(err, &notFoundErr) {
			//nolint
			c.JSON(http.StatusNotFound, RequestError{
				StatusCode: http.StatusNotFound,
				Message:    err.Error(),
			})
			return
		}

		if errors.Is(err, ErrUnauthorized) {
			//nolint
			c.JSON(http.StatusUnauthorized, RequestError{
				StatusCode: http.StatusUnauthorized,
				Message:    err.Error(),
			})
			return
		}

		// check if error is known type to be handled differently
		var myErr *ValidationError
		if errors.As(err, &myErr) {
			//nolint
			c.JSON(http.StatusBadRequest, ValidationError{
				StatusCode: http.StatusBadRequest,
				Errors:     myErr.Errors,
			})
			return
		}

		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			message, err := httpError.Message.(string)
			if !err {
				message = "Unknown (not string) error"
			}

			//nolint
			c.JSON(http.StatusBadRequest, RequestError{
				StatusCode: httpError.Code,
				Message:    message,
			})
			return
		}

		//nolint
		c.JSON(http.StatusConflict, RequestError{
			StatusCode: http.StatusConflict,
			Message:    err.Error(),
		})

		e.DefaultHTTPErrorHandler(err, c)
	}

	// Validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Global rate limiter
	if a.Options.RATE_LIMITER > 0 {
		rateMinimum := rate.Limit(a.Options.RATE_LIMITER)
		rateMaximum := a.app.Options.RATE_LIMITER * 2

		config := middleware.RateLimiterConfig{
			Skipper: middleware.DefaultSkipper,
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: rateMinimum, Burst: rateMaximum, ExpiresIn: 1 * time.Minute},
			),
			IdentifierExtractor: func(ctx echo.Context) (string, error) {
				id := ctx.RealIP()
				return id, nil
			},
			ErrorHandler: func(context echo.Context, err error) error {
				return context.JSON(http.StatusForbidden, nil)
			},
			DenyHandler: func(context echo.Context, identifier string, err error) error {
				return context.JSON(http.StatusTooManyRequests, nil)
			},
		}

		e.Use(middleware.RateLimiterWithConfig(config))
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			spew.Dump(values.Error)
			if values.Error != nil {
				msg := fmt.Sprintf("[error:%s] echo request error", values.Error.Error())

				logrus.WithFields(logrus.Fields{
					"uri":     values.URI,
					"status":  values.Status,
					"latency": values.Latency.Nanoseconds(),
					"ip":      values.RemoteIP,
				}).Error(msg)
			} else {
				msg := "request: " + values.URI

				logrus.WithFields(logrus.Fields{
					"uri":     values.URI,
					"status":  values.Status,
					"latency": values.Latency.Nanoseconds(),
					"ip":      values.RemoteIP,
				}).Info(msg)
			}

			return nil
		},
	}))

	// Routers
	initMetricsRoutes(a, e)
	initOpenAPIProfileRouters(a, e)
	initOpenAPIMainRouters(a, e)
	initOpenAPIFederationRouters(a, e)
	initOpenAPIProjectRouters(a, e)
	initOpenAPITaskRouters(a, e)
	initOpenAPIReminderRouters(a, e)
	initOpenAPIcatalogRouters(a, e)

	// Special routes
	e.File("/openapi.yaml", "./openapi.yaml", middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "pong")
	})

	e.GET("/ws", hello(a, e))

	e.GET("/seed", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")

		c.Response().WriteHeader(http.StatusOK)

		i := 0
		ch := make(chan string, 100)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("exception: %s", string(debug.Stack()))

					msg2 := fmt.Sprintf("id: %v\nevent: %s\ndata: {'msg':%s}\n\n", i, "seed", "error")
					fmt.Fprint(c.Response(), msg2)

					close(ch)
					return
				}
			}()

			usersCount := helpers.MustInt(c.QueryParam("usersCount"))
			projectsCount := helpers.MustInt(c.QueryParam("projectsCount"))
			cores := helpers.MustInt(c.QueryParam("cores"))
			tasksCountPerCore := helpers.MustInt(c.QueryParam("tasksCountPerCore"))
			batch := helpers.MustInt(c.QueryParam("batch"))

			err := a.app.Seed(ch, usersCount, projectsCount, cores, tasksCountPerCore, batch)
			if err != nil {
				logrus.Error(err)
			}
		}()

		for {
			// check chan close
			if v, ok := <-ch; ok {
				msg := v
				i++

				msg2 := fmt.Sprintf("id: %v\nevent: %s\ndata: {'msg':%s}\n\n", i, "seed", msg)
				fmt.Fprint(c.Response(), msg2)
				c.Response().Flush()
			} else {
				break
			}
		}

		return nil
	})

	e.GET("/seed_task", func(c echo.Context) error {
		total := helpers.MustInt(c.QueryParam("total"))
		projectUUID := uuid.MustParse(c.QueryParam("project_uuid"))
		createdBy := c.QueryParam("created_by")
		randomImplemented := c.QueryParam("random_implemented") == "true"
		commentsMax := helpers.MustInt(c.QueryParam("comments_max"))

		if total > 1000 {
			return errors.New("total must be < 1000")
		}

		dmns, err := a.app.SeedTasks(c.Request().Context(), total, projectUUID, createdBy, randomImplemented, commentsMax)
		if err != nil {
			logrus.Error(err)
			return err
		}

		err = c.JSON(http.StatusOK, dmns)

		return err
	})

	a.Router = e

	return e
}

func (a *Web) Run() {
	go func() {
		if err := a.Router.Start(fmt.Sprintf(":%d", a.Port)); err != nil && errors.Is(err, http.ErrServerClosed) {
			a.Router.Logger.Fatal("🙏 shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Router.Shutdown(ctx); err != nil {
		a.Router.Logger.Fatal(err)
	}
}
