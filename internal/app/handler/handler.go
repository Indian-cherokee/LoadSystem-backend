package handler

import (
	"web/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const hardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPI(router *gin.RouterGroup) {
	router.GET("/loads", h.GetLoads)
	router.GET("/loads/:id", h.GetLoadByID)
	router.POST("/loads", h.CreateLoad)
	router.PUT("/loads/:id", h.UpdateLoad)
	router.DELETE("/loads/:id", h.DeleteLoad)
	router.POST("/loads/:id/image", h.UploadLoadImage)
	router.POST("/load-sessions/draft/loads/:load_id", h.AddLoadToDraft)

	router.GET("/load-sessions/cart", h.GetCartBadge)
	router.GET("/load-sessions", h.GetLoadSessions)
	router.GET("/load-sessions/:id", h.GetLoadSession)
	router.PUT("/load-sessions/:id", h.UpdateLoadSession)
	router.PUT("/load-sessions/:id/form", h.FormLoadSession)
	router.PUT("/load-sessions/:id/resolve", h.ResolveLoadSession)
	router.DELETE("/load-sessions/:id", h.DeleteLoadSession)

	router.DELETE("/load-sessions/:id/loads/:load_id", h.RemoveLoadFromSession)
	router.PUT("/load-sessions/:id/loads/:load_id", h.UpdateLoadToCalculation)

	router.POST("/users", h.RegisterUser)
	router.GET("/users/:id", h.GetUserData)
	router.PUT("/users/:id", h.UpdateUserData)
	router.POST("/auth/login", h.Login)
	router.POST("/auth/logout", h.Logout)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
