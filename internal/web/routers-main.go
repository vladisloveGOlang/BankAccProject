package web

import (
	"context"
	"fmt"

	oapi "github.com/krisch/crm-backend/internal/web/omain"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func initOpenAPIMainRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oMain").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) GetAbout(_ context.Context, _ oapi.GetAboutRequestObject) (oapi.GetAboutResponseObject, error) {
	return oapi.GetAbout200JSONResponse{
		Version:   a.Version,
		BuildTime: a.BuildTime,
		Uuid:      a.UUID,
		StartedAt: a.Now,
		Tag:       a.Tag,
	}, nil
}

func (a *Web) GetHealth(context.Context, oapi.GetHealthRequestObject) (oapi.GetHealthResponseObject, error) {
	// todo: update to async
	postgres, err := a.app.HealthService.PingPostgres()
	if err != nil {
		logrus.Error(err)
	}

	redis, err := a.app.HealthService.PingRedis()
	if err != nil {
		logrus.Error(err)
	}

	return oapi.GetHealth200JSONResponse{
		Status:   "ok",
		Redis:    fmt.Sprintf("%v", redis),
		Postgres: fmt.Sprintf("%v", postgres),
	}, nil
}
