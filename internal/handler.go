package internal

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"strings"
)

//region Common handlers

//endregion

//region Router functions

func processOathkeeperRequest(c echo.Context) error {
	fmt.Println("\n##### HRBAC START #####\b")
	or := c.(*OathkeeperContext)

	fmt.Println(or.Payload)
	user := memorycache.GetUser(or.Payload.Subject, true)
	fmt.Println(user)

	if user == nil {
		return jsonError(or, 403, "User not found", nil)
	}

	if !user.Enabled {
		return jsonError(or, 403, "User is disabled", nil)
	}

	if !user.Verified {
		return jsonError(or, 403, "User hasn't verified their account", nil)
	}

	if user.DeletedAt != nil {
		return jsonError(or, 403, "User has been deleted", nil)
	}

	service := new(Service)
	if err := dbInstance.Model(&Service{}).First(service, "name", or.Payload.Service); err.Error != nil {
		return jsonError(or, 403, "Failed to find service", err.Error)
	}

	groups := []Group{}
	gids := []string{}

	if err := dbInstance.Table("user_group").Select("group_id").Where("user_id", user.ID).Find(&gids); err.Error != nil {
		return jsonError(or, 403, "Failed to find user groups", err.Error)
	}

	if len(gids) == 0 {
		return jsonError(or, 403, "Access denied", nil)
	}

	if err := dbInstance.Model(&Group{}).Where("id in ?", gids).Find(&groups); err.Error != nil {
		return jsonError(or, 403, "Failed to find groups", err.Error)
	}

	verb := new(Verb)
	if err := dbInstance.Model(&Verb{}).Where("name = ?", strings.ToUpper(or.Payload.Verb)).First(&verb); err.Error != nil {
		return jsonError(or, 403, "Failed to find verb", err.Error)
	}

	fmt.Println(fmt.Sprintf("User %s trying to access service %s via the %s method", user.FirstName, service.Name, verb.Name))

	accessGranted := false

	match := int64(0)
	if err := dbInstance.Table("service_group_verb").Where("service_id = ? AND verb_id = ? AND group_id = ?", service.ID, verb.ID, groups[0].ID).Count(&match); err.Error != nil {
		return jsonError(or, 403, "Access denied", nil)
	}

	if match == 0 {
		return jsonError(or, 403, "Access denied", nil)
	}

	accessGranted = true

	/**
	This is the logic that we need to process:
	1) The request comes in. We collect all of the following:
	  1.1) Service - make sure it exists ✓
	  1.2) Model - make sure it exists in the database (NOT YET IMPLEMENTED)
	  1.3) Verb - make sure its within valid ranges ✓
	2) We fetch user's group(s) ✓
	  2.1) If no groups are found, flow is terminated and access is denied ✓
	  2.2) If group has a parent group, repeat steps 2.2 onwards until flow is terminated for each parent group (NOT YET IMPLEMENTED)
	3) We fetch the group's service action for that verb
	  3.1) If the action is deny, the flow is terminated and access is denied
	  3.2) If there are no matches found or the action is allowed, we evaluate model access
	  3.2.1) Fetch service model for the group and action
	  3.2.2) If no result is found or the action is deny, terminate the flow and deny access
	*/

	fmt.Println("\n##### HRBAC END #####\b")

	if accessGranted {
		return or.JSON(200, struct {
			Message string `json:"message"`
		}{
			Message: "Hello, world",
		})
	}
	return jsonError(or, 403, "Access denied", nil)
}

//endregion
