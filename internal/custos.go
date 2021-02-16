package internal

import (
	"encoding/json"
	"github.com/Entrio/subenv"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"time"
)

func GetCustosIdentities(c echo.Context) error {
	return nil
}

func FetchKratosIdentities(url string) error {

	client := http.Client{
		Timeout: time.Second * 3,
	}

	request, err := http.NewRequest("GET", subenv.Env("KRATOS_URL", "http://192.168.2.9:4434/identities"), nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)

	if err != nil {
		return err
	}

	bufferBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	kratosUser := new([]KratosUser)
	if err := json.Unmarshal(bufferBytes, kratosUser); err != nil {
		return err
	}

	if err := ProcessUsers(kratosUser); err != nil {
		return err
	}

	return nil
}
