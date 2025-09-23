package api

import (
	"RIP/internal/app/handler"
	"RIP/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start!")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/resources", "./resources")

	r.GET("/loads", handler.GetLoads)
	r.GET("/load/:id", handler.GetLoad)
	r.GET("/calc", handler.GetCalc)

	r.Run()

	log.Println("Server terminated!")
}
