package handler

import (
	"web/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/loads", h.GetAllLoads)
	router.GET("/load/:id", h.GetLoadByID)
	router.GET("/load_calculation/:load_session_id", h.GetLoadSession)
	router.POST("/load_calculation/add/load/:load_id", h.AddLoadToLoadSession)
	router.POST("/load_calculation/:load_session_id/delete", h.DeleteLoadSession)

}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
