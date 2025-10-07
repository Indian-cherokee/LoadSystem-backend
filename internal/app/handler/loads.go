package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllLoads(ctx *gin.Context) {
	var loads []ds.Loads
	var err error

	searchingLoads := ctx.Query("searchingLoads")
	if searchingLoads == "" {
		loads, err = h.Repository.GetAllLoads()
	} else {
		loads, err = h.Repository.SearchLoadsByName(searchingLoads)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	draftLoadSession, err := h.Repository.GetDraftLoadSession(hardcodedUserID)
	var load_sessionID uint = 0
	var loadsCount int = 0

	if err == nil && draftLoadSession != nil {
		fullLoadSession, err := h.Repository.GetLoadSessionWithLoads(draftLoadSession.ID)
		if err == nil {
			load_sessionID = fullLoadSession.ID
			loadsCount = len(fullLoadSession.LoadsLink)
		}
	}

	ctx.HTML(http.StatusOK, "loads.html", gin.H{
		"loads":          loads,
		"loadsSearch":    searchingLoads,
		"load_sessionID": load_sessionID,
		"loadsCount":     loadsCount,
	})
}

func (h *Handler) GetLoadByID(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	load, err := h.Repository.GetLoadByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "load.html", load)
}
