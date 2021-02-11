package internal

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
	}

	// Register routes

	return e
}
