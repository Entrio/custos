package internal

import (
	"github.com/labstack/echo/v4"
)

func RegisterMiddleware(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
	}

	// Register middleware

	return e
}
