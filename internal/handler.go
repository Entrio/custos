package internal

import "github.com/labstack/echo/v4"

//region Common handlers

func ProcessUsers(users *[]KratosUser) error {
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

	return or.JSON(200, struct {
		Message string `json:"message"`
	}{
		Message: "Hello, world",
	})
}

//endregion
