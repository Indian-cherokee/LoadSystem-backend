package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCartBadge(c *gin.Context) {
	draft, err := h.Repository.GetDraftLoadSession(hardcodedUserID)
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

func (h *Handler) GetLoadSessions(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	sessions, err := h.Repository.GetLoadSessionsFiltered(status, from, to)
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
	}

	// Если сессия завершена, добавляем результат расчета
	if session.Status == ds.StatusCompleted {
		totalLoad, err := h.Repository.CalculateTotalLoad(session.ID)
		if err == nil {
			sessionDTO.TotalLoad = &totalLoad
		}
	}

	c.JSON(http.StatusOK, sessionDTO)
}

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

func (h *Handler) FormLoadSession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormLoadSession(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия сформирована",
	})
}

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

	moderatorID := uint(hardcodedUserID)
	if err := h.Repository.ResolveLoadSession(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Сессия обработана модератором",
	})
}

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
