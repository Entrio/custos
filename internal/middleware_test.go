package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	validEmptyJson = `{}`
)

func initEchoTest() *echo.Echo {
	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	return RegisterRoutes(e)
}

func TestEmptyPayload(t *testing.T) {
	e := initEchoTest()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validEmptyJson))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"message\":\"Payload failed validation\"}\n")

}

func TestInvalidMethod(t *testing.T) {
	e := initEchoTest()

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(validEmptyJson))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"message\":\"Method Not Allowed\"}\n")
}

func TestPayloadVerbMissing(t *testing.T) {
	e := initEchoTest()

	msg := `{"subject":"","service":"","model":"","id":"","action":""}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"message\":\"Payload failed validation\"}\n")

}

func TestPayloadWrongVerb(t *testing.T) {
	e := initEchoTest()

	msg := `{"verb":"get","subject":"","service":"","model":"","id":"","action":""}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"message\":\"Payload failed validation\"}\n")

}

func TestValidPayload(t *testing.T) {
	e := initEchoTest()

	msg := `{"verb":"get","subject":"","service":"medical","model":"drug","id":"","action":""}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"message\":\"Payload failed validation\"}\n")

}
