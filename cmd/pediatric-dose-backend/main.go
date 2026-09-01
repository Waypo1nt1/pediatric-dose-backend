package main

import (
	"log"

	"pediatric-dose-backend/internal/api"
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
