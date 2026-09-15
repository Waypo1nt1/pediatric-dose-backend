package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/dsn"
	"pediatric-dose-backend/internal/app/handler"
	"pediatric-dose-backend/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	drugRepository, err := repository.NewRepository(dsn.FromEnv())
	if err != nil {
		logrus.Error("ошибка подключения к базе данных препаратов: ", err)
		return
	}

	drugHandler := handler.NewHandler(drugRepository)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/drugs", drugHandler.GetDrugCatalog)
	r.POST("/drugs/delete", drugHandler.DeleteDrug)
	r.GET("/drug-feed/*drug_id", drugHandler.GetDrugFeed)
	r.GET("/drug-draft", drugHandler.GetDrugDraft)
	r.POST("/drug-draft", drugHandler.CreateDrugDraft)
	r.POST("/drug-draft/publish", drugHandler.PublishDrugDraft)

	if err := r.Run("localhost:8080"); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
