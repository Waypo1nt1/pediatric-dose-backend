package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	if err := r.Run("localhost:8080"); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
