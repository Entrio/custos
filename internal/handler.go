package internal

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

//region Common handlers

//endregion

//region Router functions

func processOathkeeperRequest(c echo.Context) error {

	or := c.(*OathkeeperContext)

	fmt.Println(or.Payload)

	user := memorycache.GetUser(or.Payload.Subject)

	if user == nil {
		return or.JSON(403, nil)
	}

	//TODO: Check fot banned users

	return or.JSON(200, struct {
		Message string `json:"message"`
	}{
		Message: "Hello, world",
	})
}

//endregion
