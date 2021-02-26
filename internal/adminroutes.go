package internal

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func registerAdminRoutes(e *echo.Echo) *echo.Echo {

	e.GET("identities", getIdentities)
	e.PUT("identities/:id", updateIdentity)

	e.GET("groups", getGroups)
	e.POST("groups", addGroup)
	e.PUT("groups/:id", updateGroup)
	e.POST("groups/:id/members/delete", deleteGroupMember)
	e.POST("groups/:id/members/add", addGroupMembers)
	e.DELETE("groups/:id", deleteGRoup)

	e.GET("services", getServices)
	e.POST("services", addService)
	e.PUT("services/:id", updateServiceGroups)

	return e
}

func getIdentities(c echo.Context) error {
	idents := new([]User)
	dbInstance.Find(idents)

	return c.JSON(200, idents)
}

//region Groups

func getGroups(c echo.Context) error {
	groups := new([]Group)
	dbInstance.Preload("ParentGroup").Preload("Users").Find(groups)

	return c.JSON(200, groups)
}

func updateIdentity(c echo.Context) error {

	id := c.Param("id")
	user := new(User)
	result := dbInstance.Where("id = ?", id).First(user)

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

	res := dbInstance.Create(group)

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

func updateGroup(c echo.Context) error {

	type r struct {
		Enabled     *bool
		Name        *string `json:"newName" validate:"omitempty,alphanum,max=255,min=3"`
		Description *string `json:"newDescription" validate:"omitempty,max=255,min=3"`
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

	group := new(Group)
	result := dbInstance.Model(&Group{}).Where(&Base{ID: uuid.FromStringOrNil(c.Param("id"))}).First(group)

	if group.Protected {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(403, struct {
				Message string `json:"message"`
			}{
				Message: "Cannot alter a protected group",
			})
		}
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.JSON(404, nil)
	}

	changed := false

	if incoming.Name != nil {
		if group.Name != *incoming.Name {
			changed = true
			group.Name = *incoming.Name
		}
	}

	if incoming.Description != nil {
		if *group.Description != *incoming.Description {
			changed = true
			group.Description = incoming.Description
		}
	}

	if incoming.Enabled != nil {
		if group.Enabled != *incoming.Enabled {
			changed = true
			group.Enabled = *incoming.Enabled
		}
	}

	if changed {
		group.UpdatedAt = time.Now()
		dbInstance.Save(group)
	}

	dbInstance.Model(group).Association("Users").Find(&group.Users)

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

	result := dbInstance.Model(
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

	assoc := []map[string]interface{}{}

	for _, k := range *users {
		assoc = append(assoc, map[string]interface{}{
			"user_id":    k.ID,
			"group_id":   c.Param("id"),
			"created_at": time.Now(),
			"updated_at": time.Now(),
		})
	}

	result := dbInstance.Table("user_group").Clauses(clause.OnConflict{DoNothing: true}).Create(assoc)

	/*
		result := dbInstance.Debug().Model(
				&Group{
					Base: Base{
						ID: uuid.FromStringOrNil(c.Param("id")),
					},
				}).Association("Users").Append(users)

	*/

	if result.Error != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: result.Error.Error(),
		})
	}

	return c.JSON(200, users)
}

func deleteGRoup(c echo.Context) error {
	group := new(Group)

	if err := dbInstance.Model(&Group{}).First(group, "id", c.Param("id")); err.Error != nil {
		return jsonError(c, 400, "This group doesnt exist", err.Error)
	}

	// Delete service associations
}

//endregion

//region Services

func getServices(c echo.Context) error {
	services := []Service{}

	res := dbInstance.Model(&Service{}).Preload("Verbs").Find(&services)

	if res.Error != nil {
		return c.JSON(500, struct {
			Message string `json:"message"`
		}{
			Message: "Failed to fetch data",
		})
	}

	sgv := new([]ServiceGroupVerbs)
	ids := []string{}

	for _, v := range services {
		ids = append(ids, v.ID.String())
	}

	res2 := dbInstance.Table("service_group_verb").Where("service_id IN ?", ids).Find(sgv)

	if res2.Error != nil {
		return c.JSON(500, struct {
			Message string `json:"message"`
		}{
			Message: "Failed to fetch group data",
		})
	}

	for k, s := range services {
		g := []ServiceGroupVerbs{}

		for _, gv := range *sgv {
			if gv.ServiceID == s.ID {
				g = append(g, gv)
			}
		}

		services[k].GroupVerbs = g
	}

	return c.JSON(200, services)
}

func addService(c echo.Context) error {
	type b struct {
		Name        string   `json:"name" validate:"alphanum,max=255,min=3"`
		Description string   `json:"description" validate:"max=255,min=3"`
		Verbs       []string `json:"verbs" validate:"required,dive,oneof=POST GET PUT DELETE"`
	}

	newService := new(b)

	if err := c.Bind(newService); err != nil {
		fmt.Println(err.Error())
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Invalid payload given",
		})
	}

	if err := c.Validate(newService); err != nil {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Payload failed validation",
		})
	}

	matchCount := int64(0)
	dbInstance.Model(&Service{}).Where("name = ?", newService.Name).Count(&matchCount)

	if matchCount > 0 {
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Service with that name exists",
		})
	}

	verbs := new([]Verb)
	dbInstance.Model(&Verb{}).Where("name in ?", newService.Verbs).Find(verbs)
	if len(*verbs) != len(newService.Verbs) {
		fmt.Println(fmt.Sprintf("Got %d verbs in request, found %d in DB", len(newService.Verbs), len(*verbs)))
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: "Unknown verbs",
		})
	}

	tx := dbInstance.Begin()

	service := &Service{
		Base: Base{
			ID: uuid.NewV4(),
		},
		Name:        newService.Name,
		Description: newService.Description,
	}

	res := tx.Create(service)

	if res.Error != nil {
		tx.Rollback()
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: res.Error.Error(),
		})
	}

	assoc := []map[string]interface{}{}

	for _, k := range *verbs {
		assoc = append(assoc, map[string]interface{}{
			"service_id": service.ID,
			"verb_id":    k.ID,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		})
	}

	result := tx.Table("service_verb").Clauses(clause.OnConflict{DoNothing: true}).Create(assoc)

	if result.Error != nil {
		tx.Rollback()
		return c.JSON(400, struct {
			Message string `json:"message"`
		}{
			Message: res.Error.Error(),
		})
	}

	tx.Commit()

	dbInstance.Model(service).Association("Verbs").Find(&service.Verbs)

	memorycache.AddItem(fmt.Sprintf("s_%s", service.ID), service, nil)

	return c.JSON(200, service)
}

func updateServiceGroups(c echo.Context) error {
	type req struct {
		GroupVerb []string `json:"groups"`
	}

	id := c.Param("id")

	groupVerbs := new(req)

	if err := c.Bind(groupVerbs); err != nil {
		return jsonError(c, 400, "Failed to bind to body", err)
	}

	if err := c.Validate(groupVerbs); err != nil {
		return jsonError(c, 400, "Failed to validate body", err)
	}

	// Need to parse each element of the slice and do the following:
	// 0) If there are no elements, clear all of the groups for current service. Aka delete the relations
	// 1) Validate group on the left side
	// 2) Validate verb on the right side
	// 3) Insert into the relation table correct data
	// 4) respond with the update relations

	// 0) Check for empty
	if len(groupVerbs.GroupVerb) == 0 {
		// we need to clear assocs for this group
		tx := dbInstance.Begin()
		tx.Exec("DELETE FROM service_group_verb WHERE service_id = ?", id)
		//TODO: Remove from cache
		tx.Commit()
	}

	serviceGroupVerb := []map[string]interface{}{}
	for _, v := range groupVerbs.GroupVerb {
		res := strings.Split(v, ":")
		serviceGroupVerb = append(serviceGroupVerb, map[string]interface{}{
			"service_id": id,
			"group_id":   res[0],
			"verb_id":    res[1],
			"created_at": time.Now(),
			"updated_at": time.Now(),
		})
	}

	tx := dbInstance.Begin()
	tx.Exec("DELETE FROM service_group_verb WHERE service_id = ?", id)
	if res := tx.Table("service_group_verb").Clauses(clause.OnConflict{DoNothing: true}).Create(&serviceGroupVerb); res.Error != nil {
		return jsonError(c, 400, "Failed to bind groups to service", res.Error)
		tx.Rollback()
	}

	tx.Commit()

	service := new(Service)
	dbInstance.Model(&Service{}).Preload("Verbs").First(service, "id", id)
	dbInstance.Table("service_group_verb").Where("service_id", id).Find(&service.GroupVerbs)

	return c.JSON(200, service)
}

//endregion
