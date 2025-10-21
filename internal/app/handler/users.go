package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) RegisterUser(c *gin.Context) {
	var req ds.UserRegisterRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	user := ds.Users{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Moderator: false,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	userDTO := ds.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Moderator: user.Moderator,
	}

	c.JSON(http.StatusCreated, userDTO)
}

func (h *Handler) GetUserData(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	userDTO := ds.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Moderator: user.Moderator,
	}
	c.JSON(http.StatusOK, userDTO)
}

func (h *Handler) UpdateUserData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.UserUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Данные пользователя обновлены",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req ds.UserLoginRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Username)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	response := ds.LoginResponse{
		Token: "wanderful",
		User: ds.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Moderator: user.Moderator,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, ds.SuccessResponse{
		Message: "Деавторизация прошла успешно",
	})
}
