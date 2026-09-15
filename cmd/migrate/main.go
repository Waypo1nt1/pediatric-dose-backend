package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pediatric-dose-backend/internal/app/ds"
	"pediatric-dose-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Drug{},
		&ds.DrugLike{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
