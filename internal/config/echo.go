package config

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func NewEcho(config *viper.Viper) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Debug = true

	// Set custom error handler
	e.HTTPErrorHandler = NewErrorHandler()

	return e
}

// NewErrorHandler returns a custom Echo HTTP error handler
func NewErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		msg := err.Error()

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if he.Message != nil {
				msg = he.Message.(string)
			}
		}

		// Send JSON response
		if !c.Response().Committed {
			c.JSON(code, map[string]string{
				"errors": msg,
			})
		}
	}
}
