package internal

import (
	"fmt"
	"github.com/Entrio/subenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"strings"
)

func RegisterRoutes(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
		fmt.Println("New echo from RegisterRoutes")
	}

	origins := []string{"http://127.0.0.1:8080", "http://192.168.1.33:8080"}

	if subenv.Env("ALLOWED_CORS", "") != "" {
		origins = strings.Split(subenv.Env("ALLOWED_CORS", ""), "|")
	}

	log.Println("[PUB] Allowed CORS domains:", origins)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     origins,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials},
	}))

	// Register routes
	e.POST("*", processOathkeeperRequest, validateOathkeeperRequest)

	return e
}

func RegisterAdminRoutes(e *echo.Echo) *echo.Echo {
	if e == nil {
		e = echo.New()
		fmt.Println("New echo from RegisterAdminRoutes")
	}

	origins := []string{"http://127.0.0.1:8080", "http://192.168.1.33:8080"}

	if subenv.Env("ADMIN_ALLOWED_CORS", "") != "" {
		origins = strings.Split(subenv.Env("ADMIN_ALLOWED_CORS", ""), "|")
	}

	log.Println("[ADM] Allowed CORS domains:", origins)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     origins,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials},
	}))

	registerAdminRoutes(e)

	return e
}
