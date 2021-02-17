package internal

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
		fmt.Println("New echo from RegisterRoutes")
	}

	// Register routes
	e.POST("*", processOathkeeperRequest, validateOathkeeperRequest)

	return e
}

func RegisterAdminRoutes(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
		fmt.Println("New echo from RegisterAdminRoutes")
	}

	registerAdminRoutes(e)

	return e
}
