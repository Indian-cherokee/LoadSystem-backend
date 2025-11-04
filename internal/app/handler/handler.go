package handler

import (
	"web/internal/app/config"
	"web/internal/app/redis"
	"web/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redisClient *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redisClient,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {

	// Доступны всем
	r.POST("/users", h.RegisterUser)
	r.POST("/auth/login", h.Login)
	r.GET("/loads", h.GetLoads)
	r.GET("/loads/:id", h.GetLoadByID)

	// Эндпоинты, доступные только авторизованным пользователям
	auth := r.Group("/")
	auth.Use(h.AuthMiddleware)
	{
		// Пользователи
		auth.POST("/auth/logout", h.Logout)
		auth.GET("/users/:id", h.GetUserData)
		auth.PUT("/users/:id", h.UpdateUserData)
		// Сессии загрузок
		auth.POST("/load-sessions/draft/loads/:load_id", h.AddLoadToDraft)
		auth.GET("/load-sessions/cart", h.GetCartBadge)
		auth.GET("/load-sessions", h.GetLoadSessions)
		auth.GET("/load-sessions/:id", h.GetLoadSession)
		auth.PUT("/load-sessions/:id", h.UpdateLoadSession)
		auth.PUT("/load-sessions/:id/form", h.FormLoadSession)
		auth.DELETE("/load-sessions/:id", h.DeleteLoadSession)
		auth.DELETE("/load-sessions/:id/loads/:load_id", h.RemoveLoadFromSession)
		auth.PUT("/load-sessions/:id/loads/:load_id", h.UpdateLoadToCalculation)
	}

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Управление нагрузками (создание, изменение, удаление)
		moderator.POST("/loads", h.CreateLoad)
		moderator.PUT("/loads/:id", h.UpdateLoad)
		moderator.DELETE("/loads/:id", h.DeleteLoad)
		moderator.POST("/loads/:id/image", h.UploadLoadImage)

		// Управление сессиями (завершение/отклонение)
		moderator.PUT("/load-sessions/:id/resolve", h.ResolveLoadSession)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
