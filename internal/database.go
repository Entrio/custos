package internal

import (
	"fmt"
	"github.com/Entrio/subenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

/**
https://app.dbdesigner.net/designer/schema/393779
*/

var (
	dbInstance  *gorm.DB
	memorycache *MemoryCache
)

func InitializeDB() (func(), error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d TimeZone=%s",
		subenv.Env("DB_HOST", "192.168.2.9"),
		subenv.Env("DB_USER", "postgres"),
		subenv.Env("DB_PASSWORD", "12345Aa"),
		subenv.Env("DB_NAME", "custos"),
		subenv.EnvI("DB_PORT", 5432),
		subenv.Env("DB_TZ", "Asia/Almaty"),
	)

	var err error
	dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if db, dberr := dbInstance.DB(); dberr != nil {
		return nil, dberr
	} else {
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(50)
		db.SetConnMaxLifetime(time.Minute * 30)
	}

	fmt.Println(
		fmt.Sprintf("Connected to %s on port %d",
			subenv.Env("DB_HOST", "192.0168.2.9"),
			subenv.EnvI("DB_PORT", 5432),
		),
	)

	memorycache = NewMemoryCache()

	if subenv.EnvB("APP_DB_DEBUG", true) {
		dbInstance = dbInstance.Debug()
	}

	migrate()

	return func() {
		fmt.Println("Closing link to database")
	}, nil
}

func migrate() {
	fmt.Println("Performing migrations...")
	dbInstance.AutoMigrate(&Verb{}, &Service{}, &User{})
	fmt.Println("PMigrations complete...")
}
