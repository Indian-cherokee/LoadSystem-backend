package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины (авторизованный пользователь)
// @Description  Возвращает ID черновика текущего пользователя и количество нагрузок в нем.
// @Tags         load-sessions
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.CartBadgeDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/cart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	draft, err := h.Repository.GetDraftLoadSession(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			LoadSessionID: nil,
			LoadsCount:    0,
		})
		return
	}

	fullSession, err := h.Repository.GetLoadSessionWithLoads(draft.ID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			LoadSessionID: nil,
			LoadsCount:    0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		LoadSessionID: &fullSession.ID,
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
			ID:          session.ID,
			Status:      session.Status,
			CreatedAt:   session.CreatedAt,
			CreatorID:   session.CreatorID,
			ModeratorID: nil,
			RoomType:    session.RoomType,
			TotalLoad:   nil, // По умолчанию null, если не рассчитано
		}

		if session.ModeratorID != nil {
			sessionDTO.ModeratorID = session.ModeratorID
		}

		// Если сессия завершена, рассчитываем total_load
		if session.Status == ds.StatusCompleted {
			totalLoad, err := h.Repository.CalculateTotalLoad(session.ID)
			if err == nil {
				sessionDTO.TotalLoad = &totalLoad
			}
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
		ID:          session.ID,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
		CreatorID:   session.CreatorID,
		ModeratorID: session.ModeratorID,
		RoomType:    session.RoomType,
		Loads:       loads,
		TotalLoad:   nil, // По умолчанию null, если не рассчитано
	}

	if session.Status == ds.StatusCompleted {
		totalLoad, err := h.Repository.CalculateTotalLoad(session.ID)
		if err == nil {
			sessionDTO.TotalLoad = &totalLoad
		}
		// Если ошибка при расчете, остается nil
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

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия обработана модератором",
	})
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
