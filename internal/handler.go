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
	fmt.Println(user)

	if user == nil {
		return or.JSON(403, nil)
	}

	if !user.Enabled {
		return or.JSON(403, struct {
			Message string `json:"message"`
		}{
			Message: "User is disabled",
		})
	}

	if !user.Verified {
		return or.JSON(403, struct {
			Message string `json:"message"`
		}{
			Message: "User has not verified their account",
		})
	}

	//TODO: Build the hrbac rules

	/**
	This is the logic that we need to process:
	1) The request comes in. We collect all of the following:
	  1.1) Service - make sure it exists
	  1.2) Model - make sure it exists in the database
	  1.3) Verb - make sure its within valid ranges
	2) We fetch user's group
	  2.1) If no groups are found, flow is terminated and access is denied
	  2.2) If group has a parent group, repeat steps 2.2 onwards until flow is terminated for each parent group
	3) We fetch the group's service action for that verb
	  3.1) If the action is deny, the flow is terminated and access is denied
	  3.2) If there are no matches found or the action is allowed, we evaluate model access
	  3.2.1) Fetch service model for the group and action
	  3.2.2) If no result is found or the action is deny, terminate the flow and deny access
	*/

	return or.JSON(200, struct {
		Message string `json:"message"`
	}{
		Message: "Hello, world",
	})
}

//endregion
