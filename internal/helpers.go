package internal

import (
	"fmt"
	"github.com/Entrio/subenv"
	"github.com/labstack/echo/v4"
)

func jsonError(c echo.Context, code int, message string, err error) error {
	if err != nil {

		if subenv.EnvB("APP_DEBUG", true) {
			fmt.Println(fmt.Sprintf("Returning error to client: %s", err.Error()))
		}

		return c.JSON(code, struct {
			Message string `json:"message"`
			Error   string `json:"error,omitempty"`
		}{
			Message: message,
			Error:   err.Error(),
		})
	}

	return c.JSON(code, struct {
		Message string `json:"message"`
	}{
		Message: message,
	})

}
