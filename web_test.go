package web

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/web"
	"github.com/krisch/crm-backend/pkg/redis"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func Readln(r *bufio.Reader) (s string, err error) {
	isPrefix := true

	var line, ln []byte
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func TestNewWeb(t *testing.T) {
	ctx := context.Background()
	file := "requests.txt"

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	//
	opt := configs.NewConfigsFromEnv()

	w := web.NewWeb(*opt)

	rds, err := redis.New(opt.REDIS_CREDS)
	if err != nil {
		logrus.Error(err)
	}

	go w.Work(ctx, rds)

	w.Init()

	//

	go w.Work(context.TODO(), rds)

	time.Sleep(2 * time.Second)

	r := bufio.NewReader(f)
	s, err := Readln(r)
	for err == nil {

		if s == "" {
			continue
		}

		// split request and response
		split := strings.Split(s, "\t")
		if len(split) != 4 {
			t.Errorf("error spliting request and response")
		}

		method := split[0]
		url := split[1]
		body := split[2]
		token := split[3]

		e := w.Router

		req := httptest.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderCookie, "TOKEN="+token)

		rec := httptest.NewRecorder()

		e.NewContext(req, rec)

		e.ServeHTTP(rec, req)

		time.Sleep(50 * time.Millisecond)

		logrus.Info(url, " ", rec.Code, " ")

		assert.NotEqual(t, http.StatusServiceUnavailable, rec.Code)

		s, err = Readln(r)

	}
}
