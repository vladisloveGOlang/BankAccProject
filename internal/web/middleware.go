package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/krisch/crm-backend/internal/app"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	logs "github.com/krisch/crm-backend/internal/logs"
	omain "github.com/krisch/crm-backend/internal/web/omain"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidAuthHeader = errors.New("ошибка получения данных пользователя")
	ErrUnauthorized      = errors.New("необходимо авторизоваться")
)

type key int

const (
	claimsKey key = iota
)

func LogMiddleware(w *app.App) func(c echo.Context, reqBody, resBody []byte) {
	return func(c echo.Context, reqBody, resBody []byte) {
		t := helpers.NewTime()

		req := c.Request()
		res := c.Response()

		if w.Options.LOG_TO_DB {
			requestData := make(map[string]interface{})
			if req.Header.Get("Content-Type") == "application/json" && len(reqBody) > 0 {
				err := json.Unmarshal(reqBody, &requestData)
				if err != nil {
					logrus.WithField("middleware", "LogMiddleware").Error(err)
				}
			}
			requestDataJSON, err := json.Marshal(requestData)
			if err != nil {
				logrus.WithField("middleware", "LogMiddleware").Error(err)
				requestDataJSON = []byte("{}")
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			Token := ""
			cookie, err := c.Cookie("TOKEN")
			if err == nil {
				Token = cookie.Value
			}

			log := logs.Log{
				HeaderXRequestID: id,
				BackendUUID:      w.Name,
				IP:               c.RealIP(),
				Host:             req.Host,
				Method:           req.Method,
				RequestURI:       req.RequestURI,
				Status:           res.Status,
				Agent:            req.UserAgent(),
				Referer:          req.Referer(),
				Start:            t.Finish().GetStart(),
				Stop:             t.Finish().GetEnd(),
				Request:          string(requestDataJSON),
				Token:            Token,
			}

			err = w.LogService.InsertLog(log)
			if err != nil {
				logrus.WithField("middleware", "LogMiddleware").Error(err)
			}
		}

		// replace uuid in the requesturi
		m1 := regexp.MustCompile("[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}")
		m2 := regexp.MustCompile(`\d+`)
		requestPath := m1.ReplaceAllString(req.RequestURI, "{UUID}")
		requestPath = m2.ReplaceAllString(requestPath, "{ID}")

		w.MetricsCounters.RequestHistogram.WithLabelValues(req.Method, requestPath).Observe(t.Secondsf())
		w.MetricsCounters.RequestHistogram.WithLabelValues(req.Method, req.Method).Observe(t.Secondsf())
	}
}

func ValidateStructMiddeware(f omain.StrictHandlerFunc, operationID string) omain.StrictHandlerFunc {
	return func(ctx echo.Context, i interface{}) (interface{}, error) {
		errs, ok := helpers.ValidationStruct(i)
		if !ok {
			logrus.Debugf("[operation:%v] validation error", operationID)
			return nil, &ValidationError{
				http.StatusBadRequest,
				errs,
			}
		}

		return f(ctx, i)
	}
}

func AuthMiddeware(a *app.App, operationIDs []string) func(f omain.StrictHandlerFunc, operationID string) omain.StrictHandlerFunc {
	return func(f omain.StrictHandlerFunc, operationID string) omain.StrictHandlerFunc {
		return func(ctx echo.Context, i interface{}) (interface{}, error) {
			// check operationID route  in operationIDs
			if len(operationIDs) > 1 && !helpers.InArray(operationID, operationIDs) {
				return f(ctx, i)
			}

			logrus.Debugf("[operation:%v] auth middleware", operationID)

			// @todo: hack
			tracer := ctx.Get("tracer")

			ctx, err := checkAuth(ctx, operationID, a.JWT)
			if err != nil {
				return nil, ErrUnauthorized
			}

			//nolint // it is @todo: hack
			ctx = echo.New().NewContext(ctx.Request().Clone(context.WithValue(ctx.Request().Context(), "tracer", tracer)), ctx.Response())

			return f(ctx, i)
		}
	}
}

func checkAuth(ctx echo.Context, operationID string, j jwt.IJWT) (echo.Context, error) {
	cookie, err := ctx.Cookie("TOKEN")

	var claims jwt.Claims

	if err == nil {
		logrus.Debugf("[operation:%v] auth with cookie", operationID)
		tokens := strings.Split(cookie.Value, "&")
		if len(tokens) != 2 {
			return nil, errors.New("invalid cookie token")
		}

		access, refresh := tokens[0], tokens[1]

		claims, err = j.ParseJWT(access)

		if errors.Is(err, jwt.ErrTokenExpired) {
			logrus.Debugf("[operation:%v] try to refresh token", operationID)

			access, expAt, err := j.RefreshAccessToken(refresh)
			if err != nil {
				return nil, err
			}

			ctx.SetCookie(j.GenerateTokenCookie(access, refresh, expAt))

			logrus.Debugf("[operation:%v] token cookie refreshed", operationID)

			claims, err = j.ParseJWT(access)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	} else {
		logrus.Debugf("[operation:%v] auth with bearer", operationID)

		authHdr := ctx.Request().Header.Get("Authorization")

		if len(authHdr) < 10 {
			return nil, errors.New("no authorization header")
		}

		prefix := "Bearer "
		if !strings.HasPrefix(authHdr, prefix) {
			return nil, ErrInvalidAuthHeader
		}

		token := strings.TrimPrefix(authHdr, prefix)

		claims, err = j.ParseJWT(token)
		if err != nil {
			return nil, err
		}
	}

	ctx = echo.New().NewContext(ctx.Request().Clone(context.WithValue(ctx.Request().Context(), claimsKey, claims)), ctx.Response())

	return ctx, nil
}
