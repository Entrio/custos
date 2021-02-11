package main

/**
https://github.com/go-pg/pg
*/

import (
	"fmt"
	"github.com/Entrio/custos/internal"
	"os"
	"os/signal"
)

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	dbClose := internal.InitializeDB()

	// Fetch the records from kratos and start cronjob

	e := internal.RegisterRoutes(internal.RegisterMiddleware(nil))

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			dbClose()
			os.Exit(1)
		}
	}()

	e.Logger.Fatal(e.Start(":1323"))
}
