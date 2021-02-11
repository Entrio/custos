package internal

import (
	"fmt"
	"github.com/go-pg/pg/v10"
)

var (
	dbInstance *pg.DB
)

func InitializeDB() func() {
	dbInstance = pg.Connect(&pg.Options{
		Addr:               "172.30.0.9",
		User:               "custos_user",
		Password:           "12345Aa",
		Database:           "custos",
		ReadTimeout:        1,
		WriteTimeout:       1,
		MaxRetries:         2,
		PoolSize:           10,
		MinIdleConns:       10,
		MaxConnAge:         60,
		IdleCheckFrequency: 30,
	})

	return func() {
		fmt.Println("Closing link to database")
		defer dbInstance.Close()
	}
}
