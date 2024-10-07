package web

import (
	"net/http"

	validator "github.com/go-playground/validator/v10"
	echo "github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func CommonError(err error) error {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}
