package handler

import (
	"RIP/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetLoads(ctx *gin.Context) {
	var loads []repository.Load
	var err error

	LoadsCount := 3

	searchLoad := ctx.Query("searchingLoads")
	if searchLoad == "" {
		loads, err = h.Repository.GetLoads()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		loads, err = h.Repository.GetLoadsByTitle(searchLoad)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "loads.html", gin.H{
		"loads":       loads,
		"searchLoads": searchLoad,
		"loadsCount":  LoadsCount,
	})
}

func (h *Handler) GetLoad(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	load, err := h.Repository.GetLoad(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "load.html", gin.H{
		"load": load,
	})
}

func (h *Handler) GetCalc(ctx *gin.Context) {
	page, err := h.Repository.GetCalcPage(1)
	if err != nil {
		logrus.Error(err)
	}

	type loadWithArea struct {
		Load repository.Load
		Area int
	}
	loadsInCalc := make([]loadWithArea, 0, len(page.Loads))
	for _, item := range page.Loads {
		loadsInCalc = append(loadsInCalc, loadWithArea{Load: item.Load, Area: item.Area})
	}

	ctx.HTML(http.StatusOK, "load_calculation.html", gin.H{
		"loadsInCalc": loadsInCalc,
		"roomType":    page.RoomType,
		"totalLoadKg": page.TotalLoadKg,
	})
}
