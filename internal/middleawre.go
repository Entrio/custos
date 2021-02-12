package internal

import (
	"fmt"
	"github.com/Entrio/subenv"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//region Custom oathkeeper context

type OathkeeperContext struct {
	echo.Context
	Payload *OathkeeperRequest
}

func (oc *OathkeeperContext) GetVerb() string {
	return oc.Payload.Verb
}

func (oc *OathkeeperContext) GetService() string {
	return oc.Payload.Service
}

func (oc *OathkeeperContext) GetAction() *string {
	return oc.Payload.Action
}

func (oc *OathkeeperContext) GetModel() *string {
	return oc.Payload.Model
}

func (oc *OathkeeperContext) GetModelID() *string {
	return oc.Payload.ModelID
}

//endregion

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func RegisterMiddleware(e *echo.Echo) *echo.Echo {
	if e == nil {
		fmt.Println("New echo from RegisterMiddleware")
		e = echo.New()
	}

	if subenv.EnvB("APP_DEBUG", false) {
		e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			fmt.Println(string(reqBody))
		}))
	}

	e.Use(middleware.Logger())

	e.Validator = &CustomValidator{validator: validator.New()}

	return e
}

func validateOathkeeperRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Validate the body and assign it to the context

		or := new(OathkeeperRequest)

		if err := c.Bind(or); err != nil {
			return c.JSON(403, struct {
				Message string `json:"message"`
			}{
				Message: "Invalid payload given",
			})
		}

		if err := c.Validate(or); err != nil {
			return c.JSON(403, struct {
				Message string `json:"message"`
			}{
				Message: "Payload failed validation",
			})
		}

		oc := &OathkeeperContext{
			Context: c,
			Payload: or,
		}

		return next(oc)
	}
}
