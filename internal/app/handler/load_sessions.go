package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины
// @Description  Возвращает ID черновика текущего пользователя и количество нагрузок в нем. Для неавторизованных пользователей возвращает 0.
// @Tags         load-sessions
// @Produce      json
// @Success      200 {object} ds.CartBadgeDTO
// @Router       /load-sessions/cart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, ok := getUserIDFromContextOptional(c)
	if !ok {
		// Если пользователь не авторизован, возвращаем -1 для ID сессии и 0 для количества
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			LoadSessionID: -1,
			LoadsCount:    0,
		})
		return
	}

	draft, err := h.Repository.GetDraftLoadSession(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			LoadSessionID: -1,
			LoadsCount:    0,
		})
		return
	}

	fullSession, err := h.Repository.GetLoadSessionWithLoads(draft.ID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			LoadSessionID: -1,
			LoadsCount:    0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		LoadSessionID: int(fullSession.ID),
		LoadsCount:    len(fullSession.LoadsLink),
	})
}

// GetLoadSessions godoc
// @Summary      Получить список сессий загрузок (авторизованный пользователь)
// @Description  Возвращает отфильтрованный список всех сформированных сессий (кроме черновиков и удаленных). Пользователи видят только свои сессии, модераторы - все.
// @Tags         load-sessions
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status query string false "Фильтр по статусу (draft, formed, completed, rejected)"
// @Param        from query string false "Фильтр по дате 'от' (формат YYYY-MM-DD)"
// @Param        to query string false "Фильтр по дате 'до' (формат YYYY-MM-DD)"
// @Success      200 {object} ds.PaginatedResponse
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions [get]
func (h *Handler) GetLoadSessions(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	isModerator := isUserModerator(c)

	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	sessions, err := h.Repository.GetLoadSessionsFiltered(userID, isModerator, status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var sessionDTOs []ds.LoadSessionDTO
	for _, session := range sessions {
		sessionDTO := ds.LoadSessionDTO{
			ID:             session.ID,
			Status:         session.Status,
			CreationDate:   session.CreationDate,
			CreatorID:      session.CreatorID,
			RoomType:       session.RoomType,
			ModeratorID:    nil,
			FormingDate:    session.FormingDate,
			CompletionDate: session.CompletionDate,
			TotalLoad:      session.TotalLoad, // Используем значение из БД, если есть
		}

		if session.ModeratorID != nil {
			sessionDTO.ModeratorID = session.ModeratorID
		}

		sessionDTOs = append(sessionDTOs, sessionDTO)
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: sessionDTOs,
		Total: int64(len(sessionDTOs)),
	})
}

// GetLoadSession godoc
// @Summary      Получить сессию по ID (авторизованный пользователь)
// @Description  Возвращает полную информацию о сессии, включая привязанные нагрузки.
// @Tags         load-sessions
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Success      200 {object} ds.LoadSessionDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Сессия не найдена"
// @Router       /load-sessions/{id} [get]
func (h *Handler) GetLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	session, err := h.Repository.GetLoadSessionWithLoads(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var loads []ds.LoadInSessionDTO
	for _, link := range session.LoadsLink {
		loads = append(loads, ds.LoadInSessionDTO{
			LoadID:       link.LoadID,
			LoadTitle:    link.Load.LoadTitle,
			LoadImage:    link.Load.LoadImage,
			Normative:    link.Load.Normative,
			LoadCategory: link.Load.LoadCategory,
			Area:         link.Area,
		})
	}

	sessionDTO := ds.LoadSessionDTO{
		ID:             session.ID,
		Status:         session.Status,
		CreationDate:   session.CreationDate,
		CreatorID:      session.CreatorID,
		RoomType:       session.RoomType,
		ModeratorID:    session.ModeratorID,
		FormingDate:    session.FormingDate,
		CompletionDate: session.CompletionDate,
		Loads:          loads,
		TotalLoad:      session.TotalLoad, // Используем значение из БД (рассчитывается асинхронным сервисом)
	}

	c.JSON(http.StatusOK, sessionDTO)
}

// UpdateLoadSession godoc
// @Summary      Обновить сессию (авторизованный пользователь)
// @Description  Обновляет поля сессии, доступные пользователю (например, тип помещения).
// @Tags         load-sessions
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Param        updateData body ds.LoadSessionUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/{id} [put]
func (h *Handler) UpdateLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.LoadSessionUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateLoadSessionUserFields(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Данные сессии обновлены",
	})
}

// FormLoadSession godoc
// @Summary      Сформировать сессию загрузок (авторизованный пользователь)
// @Description  Переводит черновик сессии в статус "сформирована". Требует авторизации.
// @Tags         load-sessions
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/{id}/form [put]
func (h *Handler) FormLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.FormLoadSession(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия сформирована",
	})
}

// ResolveLoadSession godoc
// @Summary      Завершить или отклонить сессию (модератор)
// @Description  Обрабатывает сессию модератором: завершает или отклоняет. Требует прав модератора.
// @Tags         load-sessions
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Param        action body ds.LoadSessionResolveRequest true "Действие (complete/reject)"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Требуются права модератора"
// @Router       /load-sessions/{id}/resolve [put]
func (h *Handler) ResolveLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.LoadSessionResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	moderatorID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.ResolveLoadSession(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// Если сессия завершена, вызываем асинхронный сервис для расчета total_load
	if req.Action == "complete" {
		go h.callAsyncService(uint(id))
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия обработана модератором",
	})
}

// callAsyncService вызывает асинхронный Django сервис для расчета total_load
func (h *Handler) callAsyncService(sessionID uint) {
	session, err := h.Repository.GetLoadSessionWithLoads(sessionID)
	if err != nil {
		logrus.Errorf("Failed to get session %d for async calculation: %v", sessionID, err)
		return
	}

	// Формируем данные для асинхронного сервиса
	loadsData := make([]map[string]interface{}, 0)
	for _, link := range session.LoadsLink {
		if link.Area == nil || *link.Area <= 0 {
			continue
		}
		loadData := map[string]interface{}{
			"area":                    *link.Area,
			"normative":               link.Load.Normative,
			"reliability_coefficient": link.Load.ReliabilityCoefficient,
			"load_category":           link.Load.LoadCategory,
		}
		loadsData = append(loadsData, loadData)
	}

	requestData := map[string]interface{}{
		"id":    sessionID,
		"loads": loadsData,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		logrus.Errorf("Failed to marshal request data for session %d: %v", sessionID, err)
		return
	}

	// URL асинхронного Django сервиса
	asyncServiceURL := "http://localhost:8000/api/calculate_total_load/"
	req, err := http.NewRequest("POST", asyncServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Errorf("Failed to create request to async service for session %d: %v", sessionID, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Failed to call async service for session %d: %v", sessionID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Async service returned error status %d for session %d", resp.StatusCode, sessionID)
		return
	}

	logrus.Infof("Async calculation started for session %d", sessionID)
}

// DeleteLoadSession godoc
// @Summary      Удалить сессию (авторизованный пользователь)
// @Description  Логически удаляет сессию (переводит в статус deleted).
// @Tags         load-sessions
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/{id} [delete]
func (h *Handler) DeleteLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteLoadSession(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия удалена",
	})
}

// RemoveLoadFromSession godoc
// @Summary      Удалить нагрузку из сессии (авторизованный пользователь)
// @Description  Удаляет нагрузку из сессии загрузок.
// @Tags         load-sessions
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Param        load_id path int true "ID нагрузки"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/{id}/loads/{load_id} [delete]
func (h *Handler) RemoveLoadFromSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	loadID, err := strconv.Atoi(c.Param("load_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveLoadFromSession(uint(sessionID), uint(loadID)); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Нагрузка удалена из сессии",
	})
}

// UpdateLoadToCalculation godoc
// @Summary      Обновить параметры нагрузки в сессии (авторизованный пользователь)
// @Description  Обновляет параметры нагрузки в сессии (например, площадь).
// @Tags         load-sessions
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID сессии"
// @Param        load_id path int true "ID нагрузки"
// @Param        updateData body ds.LoadToCalculationUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/{id}/loads/{load_id} [put]
func (h *Handler) UpdateLoadToCalculation(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	loadID, err := strconv.Atoi(c.Param("load_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.LoadToCalculationUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updateData := ds.LoadToCalculation{
		Area: req.Area,
	}

	if err := h.Repository.UpdateLoadToCalculation(uint(sessionID), uint(loadID), updateData); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Площадь нагрузки обновлена",
	})
}

// UpdateLoadSessionTotalLoad godoc
// @Summary      Обновить total_load для сессии (внутренний endpoint)
// @Description  Принимает результат расчета от асинхронного сервиса и обновляет total_load в сессии. Требует токен авторизации.
// @Tags         load-sessions
// @Accept       json
// @Param        updateData body ds.LoadSessionTotalLoadUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Неверный токен авторизации"
// @Router       /total_loads/updating [put]
func (h *Handler) UpdateLoadSessionTotalLoad(c *gin.Context) {
	// Проверка токена авторизации
	authToken := c.GetHeader("Authorization")
	const expectedToken = "loadsys12" // Псевдо-токен на 8 байт

	if authToken != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "Invalid authorization token",
		})
		return
	}

	var req ds.LoadSessionTotalLoadUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateLoadSessionTotalLoad(req.ID, req.TotalLoad); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Total load updated",
	})
}
