package api

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/handler"
	"pediatric-dose-backend/internal/app/repository"
)

const defaultMediaBaseURL = "http://localhost:9000/drug-media"

func StartServer() {
	log.Println("Server start up")

	drugRepository, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория препаратов")
		return
	}

	mediaBaseURL := os.Getenv("DRUG_MEDIA_BASE_URL")
	if mediaBaseURL == "" {
		mediaBaseURL = defaultMediaBaseURL
	}

	drugHandler := handler.NewHandler(drugRepository, mediaBaseURL)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/drugs", drugHandler.GetDrugCatalog)
	r.GET("/drug-feed/*drug_id", drugHandler.GetDrugFeed)
	r.GET("/drug-draft", drugHandler.GetDrugDraft)

	if err := r.Run("localhost:8080"); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
