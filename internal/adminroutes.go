package internal

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"time"
)

func registerAdminRoutes(e *echo.Echo) *echo.Echo {

	e.GET("identities", getIdentities)
	e.PUT("identities/:id", updateIdentity)
	e.GET("groups", getGroups)
	e.POST("groups", addGroup)
	e.POST("groups/:id/members/delete", deleteGroupMember)
	e.POST("groups/:id/members/add", addGroupMembers)

	return e
}

func getIdentities(c echo.Context) error {
	idents := new([]User)
	dbInstance.Find(idents)

	return c.JSON(200, idents)
}

func getGroups(c echo.Context) error {
	groups := new([]Group)
	dbInstance.Preload("ParentGroup").Preload("Users").Find(groups)

	return c.JSON(200, groups)
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

func addGroup(c echo.Context) error {

	type b struct {
		Name        string  `json:"name" validate:"alphanum,max=255,min=3"`
		Description string  `json:"description" validate:"max=255,min=3"`
		Parent      *string `json:"parent,omitempty" validate:"omitempty,uuid4"`
	}

	newGroup := new(b)

	if err := c.Bind(newGroup); err != nil {
		fmt.Println(err.Error())
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Invalid payload given",
		})
	}

	if err := c.Validate(newGroup); err != nil {
		fmt.Println(err.Error())
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Payload failed validation",
		})
	}

	// Check if the group name is already taken
	matchCount := int64(0)
	dbInstance.Model(&Group{}).Where("name = ?", newGroup.Name).Count(&matchCount)

	if matchCount > 0 {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Group with that name exists",
		})
	}

	if newGroup.Parent != nil {
		parent := new(Group)
		dbInstance.Where("id = ?", *newGroup.Parent).First(parent)

		if parent.Protected {
			return c.JSON(400, struct {
				Message string `json:"message"`
			}{
				Message: "Cannot inherit a protected group",
			})
		}
	}

	group := &Group{
		Base: Base{
			ID: uuid.NewV4(),
		},
		Name:          newGroup.Name,
		Description:   &newGroup.Description,
		ParentGroupID: newGroup.Parent,
	}

	res := dbInstance.Debug().Create(group)

	if res.Error != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: res.Error.Error(),
		})
	}

	memorycache.AddItem(fmt.Sprintf("g_%s", group.ID), group, nil)

	return c.JSON(200, group)
}

func deleteGroupMember(c echo.Context) error {
	type r struct {
		Users []uuid.UUID `json:"users"`
	}

	incoming := new(r)

	if err := c.Bind(incoming); err != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Invalid payload given",
		})
	}

	if err := c.Validate(incoming); err != nil {
		fmt.Println(err.Error())
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Payload failed validation",
		})
	}

	users := new([]User)
	count := int64(0)
	dbInstance.Model(&User{}).Where("id in ?", incoming.Users).Find(users).Count(&count)
	if count != int64(len(incoming.Users)) {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Unknown users",
		})
	}

	result := dbInstance.Debug().Model(
		&Group{
			Base: Base{
				ID: uuid.FromStringOrNil(c.Param("id")),
			},
		}).Association("Users").Delete(users)

	if result != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: result.Error(),
		})
	}

	return c.JSON(200, users)
}

func addGroupMembers(c echo.Context) error {
	type r struct {
		Users []uuid.UUID `json:"users"`
	}

	incoming := new(r)

	if err := c.Bind(incoming); err != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Invalid payload given",
		})
	}

	if err := c.Validate(incoming); err != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Payload failed validation",
		})
	}

	users := new([]User)
	count := int64(0)
	dbInstance.Model(&User{}).Where("id in ?", incoming.Users).Find(users).Count(&count)
	if count != int64(len(incoming.Users)) {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Unknown users",
		})
	}

	result := dbInstance.Debug().Model(
		&Group{
			Base: Base{
				ID: uuid.FromStringOrNil(c.Param("id")),
			},
		}).Association("Users").Append(users)

	if result != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: result.Error(),
		})
	}

	return c.JSON(200, users)
}
