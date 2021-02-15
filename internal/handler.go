package internal

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

//region Common handlers

func ProcessUsers(users *[]KratosUser) error {

	if users == nil {
		// this is a nullptr, do nothing
		return nil
	}

	// Add each user to cache and make sure that they exist in the database
	//expiry := time.Now().Add(time.Second * 10)
	for _, v := range *users {
		//memorycache.AddItem(v.ID, v, &expiry)
		memorycache.AddItem(v.ID, v, nil)
	}
	return nil
}

//endregion

//region Router functions

func processOathkeeperRequest(c echo.Context) error {

	or := c.(*OathkeeperContext)

	fmt.Println(or.Payload)

	user := memorycache.GetUser(or.Payload.Subject)

	if !user.VerifiableAddresses[0].Verified {
		return or.JSON(403, nil)
	}

	return or.JSON(200, struct {
		Message string `json:"message"`
	}{
		Message: "Hello, world",
	})
}

//endregion
