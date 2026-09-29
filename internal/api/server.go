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

	drugRepository, err := repository.NewRepository(dsn.FromEnv(), repository.MinioFromEnv())
	if err != nil {
		logrus.Error("ошибка подключения к базе данных препаратов: ", err)
		return
	}

	drugHandler := handler.NewHandler(drugRepository)

	r := gin.Default()

	api := r.Group("/api")

	api.GET("/drugs", drugHandler.GetDrugs)
	api.GET("/drugs/feed", drugHandler.GetDrugFeed)
	api.GET("/drugs/feed/:drug_id", drugHandler.GetDrugFeed)
	api.GET("/drugs/draft", drugHandler.GetDrugDraft)
	api.POST("/drugs", drugHandler.CreateDrug)
	api.PUT("/drugs/publish", drugHandler.PublishDrug)
	api.DELETE("/drugs/:drug_id", drugHandler.DeleteDrug)
	api.POST("/drugs/:drug_id/like", drugHandler.LikeDrug)

	api.POST("/users/register", drugHandler.RegisterUser)
	api.POST("/users/login", drugHandler.LoginUser)
	api.POST("/users/logout", drugHandler.LogoutUser)

	if err := r.Run("localhost:8080"); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
