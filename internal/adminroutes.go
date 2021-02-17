package internal

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"time"
)

func registerAdminRoutes(e *echo.Echo) *echo.Echo {

	e.GET("identities", getIdentities)
	e.PUT("identities/:id", updateIdentity)

	return e
}

func getIdentities(c echo.Context) error {
	idents := new([]User)
	dbInstance.Find(idents)

	return c.JSON(200, idents)
}

func updateIdentity(c echo.Context) error {

	id := c.Param("id")
	user := new(User)
	result := dbInstance.Debug().Where("id = ?", id).First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.JSON(404, nil)
	}

	type b struct {
		Enabled bool   `json:"enabled"`
		Reason  string `json:"reason"`
	}

	newState := new(b)

	if err := c.Bind(newState); err != nil {
		fmt.Println(err)
		return c.JSON(403, struct {
			Message string `json:"message"`
		}{
			Message: "Invalid payload given",
		})
	}

	if err := c.Validate(newState); err != nil {

		return c.JSON(403, struct {
			Message string `json:"message"`
		}{
			Message: "Payload failed validation",
		})
	}

	if !newState.Enabled && len(newState.Reason) < 1 {
		// We are blocking the user but no reason was given to why
		return c.JSON(403, struct {
			Message string `json:"message"`
		}{
			Message: "Please specify the reason why this user is being disabled",
		})
	}

	user.Enabled = newState.Enabled
	if !newState.Enabled {
		user.DisableReason = &newState.Reason
		dt := time.Now()
		user.DisabledDate = &dt
	} else {
		user.DisableReason = nil
		user.DisabledDate = nil
	}
	result = dbInstance.Save(user)

	if result.Error != nil {
		return c.JSON(400, result.Error)
	}

	memorycache.AddItem(id, user, nil)

	return c.JSON(200, id)
}
