package main

/**
https://github.com/go-pg/pg
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Entrio/custos/internal"
	"github.com/Entrio/subenv"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	CommitTag    = "#DEVEL_TAG"
	BuildDate    = "#DEVEL_DATE"
	BuildVersion = "#DEVEL_VER"
)

func main() {

	fmt.Println(
		fmt.Sprintf(
			"Launching custos...\nBuild version: %s\nBuild date: %s\nCommit tag: %s\n",
			BuildVersion, BuildDate, CommitTag,
		),
	)

	// Listen for the system kill signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// Fetch the user identities from kratos API
	users, err := fetchUserData()
	if err != nil {
		panic(err)
	}

	dbClose, err := internal.InitializeDB()
	if err != nil {
		panic(err)
	}

	if err := internal.ProcessUsers(users); err != nil {
		panic(err)
	}

	e := internal.RegisterRoutes(internal.RegisterMiddleware(nil))
	e.HideBanner = true
	e.Logger.SetLevel(log.DEBUG)

	go func() {
		e.Logger.Fatal(
			e.Start(
				fmt.Sprintf(
					"%s:%d",
					subenv.Env("LISTEN_IP", "192.168.1.33"),
					subenv.EnvI("LISTEN_PORT", 1323),
				),
			),
		)
	}()

	// Close the database link in case of abnormal program termination
	defer dbClose()

	// Listen for kill signal and terminate the application in a "friendlier" way
	select {
	case sig := <-c:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		dbClose()
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
		os.Exit(1)
	}
}

func fetchUserData() (*[]internal.KratosUser, error) {
	client := http.Client{
		Timeout: time.Second * 3,
	}

	request, err := http.NewRequest("GET", subenv.Env("KRATOS_URL", "http://192.168.2.9:4434/identities"), nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	bufferBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	// unmarshal json
	kratosUser := new([]internal.KratosUser)
	err = json.Unmarshal(bufferBytes, kratosUser)
	if err != nil {
		return nil, err
	}
	return kratosUser, nil
}
