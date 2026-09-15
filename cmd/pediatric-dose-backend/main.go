package main

import (
	"log"

	"github.com/joho/godotenv"

	"pediatric-dose-backend/internal/api"
)

func main() {
	log.Println("Application start!")
	_ = godotenv.Load()
	api.StartServer()
	log.Println("Application terminated!")
}
