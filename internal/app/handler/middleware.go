package handler

import (
	"errors"
	"net/http"
	"strings"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtPrefix    = "Bearer "
	userCtx      = "userID"
	moderatorCtx = "isModerator"
)

func (h *Handler) AuthMiddleware(c *gin.Context) {
	var tokenStr string

	// Сначала пробуем получить токен из Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, jwtPrefix) {
		tokenStr = authHeader[len(jwtPrefix):]
	} else {
		// Если нет в header, пробуем получить из Cookie
		cookieToken, err := c.Cookie("token")
		if err == nil && cookieToken != "" {
			tokenStr = cookieToken
		} else {
			h.errorHandler(c, http.StatusUnauthorized, errors.New("empty auth header or cookie"))
			c.Abort()
			return
		}
	}

	if tokenStr == "" {
		h.errorHandler(c, http.StatusUnauthorized, errors.New("token not found"))
		c.Abort()
		return
	}

	// Проверка в черном списке Redis
	err := h.Redis.CheckJWTInBlacklist(c.Request.Context(), tokenStr)
	if err == nil { // Ошибки нет -> токен найден в списке -> доступ запрещен
		h.errorHandler(c, http.StatusUnauthorized, errors.New("token is blacklisted"))
		c.Abort()
		return
	}
	if !errors.Is(err, redis.Nil) { // Если ошибка не "не найдено", а что-то другое
		h.errorHandler(c, http.StatusInternalServerError, err)
		c.Abort()
		return
	}

	// Парсинг токена
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.JWTConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		h.errorHandler(c, http.StatusUnauthorized, errors.New("invalid token"))
		c.Abort()
		return
	}

	// Сохраняем данные пользователя в контекст для дальнейшего использования
	c.Set(userCtx, claims.UserID)
	c.Set(moderatorCtx, claims.IsModerator)
	c.Next()
}

func (h *Handler) ModeratorMiddleware(c *gin.Context) {
	isModerator, exists := c.Get(moderatorCtx)
	if !exists || !isModerator.(bool) {
		h.errorHandler(c, http.StatusForbidden, errors.New("access denied: moderator rights required"))
		c.Abort()
		return
	}
	c.Next()
}

// OptionalAuthMiddleware - опциональная проверка авторизации (не прерывает выполнение, если токена нет)
func (h *Handler) OptionalAuthMiddleware(c *gin.Context) {
	var tokenStr string

	// Сначала пробуем получить токен из Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, jwtPrefix) {
		tokenStr = authHeader[len(jwtPrefix):]
	} else {
		// Если нет в header, пробуем получить из Cookie
		cookieToken, err := c.Cookie("token")
		if err == nil && cookieToken != "" {
			tokenStr = cookieToken
		}
	}

	// Если токена нет, просто продолжаем без установки контекста
	if tokenStr == "" {
		c.Next()
		return
	}

	// Проверка в черном списке Redis
	err := h.Redis.CheckJWTInBlacklist(c.Request.Context(), tokenStr)
	if err == nil { // Ошибки нет -> токен найден в списке -> не авторизуем
		c.Next()
		return
	}
	if !errors.Is(err, redis.Nil) { // Если ошибка не "не найдено", а что-то другое
		// Игнорируем ошибку Redis и продолжаем
		c.Next()
		return
	}

	// Парсинг токена
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.JWTConfig.Secret), nil
	})

	// Если токен валидный, устанавливаем контекст
	if err == nil && token.Valid {
		c.Set(userCtx, claims.UserID)
		c.Set(moderatorCtx, claims.IsModerator)
	}

	c.Next()
}
