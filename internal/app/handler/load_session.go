package handler

import (
	"errors"
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hardcodedUserID = 1

func (h *Handler) AddLoadToLoadSession(c *gin.Context) {
	loadID, err := strconv.Atoi(c.Param("load_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	load_session, err := h.Repository.GetDraftLoadSession(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newLoadSession := ds.LoadSession{
			CreatorID: hardcodedUserID,
			Status:    ds.StatusDraft,
		}
		if createErr := h.Repository.CreateLoadSession(&newLoadSession); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		load_session = &newLoadSession
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddLoadToLoadSession(load_session.ID, uint(loadID)); err != nil {
	}

	c.Redirect(http.StatusFound, "/loads")
}

func (h *Handler) GetLoadSession(c *gin.Context) {
	load_sessionID, err := strconv.Atoi(c.Param("load_session_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	load_session, err := h.Repository.GetLoadSessionWithLoads(uint(load_sessionID))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	if len(load_session.LoadsLink) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty load session page, add loads first"))
		return
	}

	c.HTML(http.StatusOK, "load_calculation.html", load_session)
}

func (h *Handler) DeleteLoadSession(c *gin.Context) {
	load_sessionID, _ := strconv.Atoi(c.Param("load_session_id"))

	if err := h.Repository.LogicallyDeleteLoadSession(uint(load_sessionID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/loads")
}
