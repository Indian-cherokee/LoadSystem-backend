package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// Поиск ID пользователя
func getUserIDFromContext(c *gin.Context) (uint, error) {
	value, exists := c.Get(userCtx)
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}

// Проверка на модератора
func isUserModerator(c *gin.Context) bool {
	value, exists := c.Get(moderatorCtx)
	if !exists {
		return false
	}

	isModerator, ok := value.(bool)
	if !ok {
		return false
	}

	return isModerator
}
