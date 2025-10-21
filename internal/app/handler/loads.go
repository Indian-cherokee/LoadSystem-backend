package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLoads(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")

	loads, err := h.Repository.GetLoadsFiltered(search, category)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var loadDTOs []ds.LoadDTO
	for _, load := range loads {
		loadDTOs = append(loadDTOs, ds.LoadDTO{
			ID:                     load.ID,
			LoadTitle:              load.LoadTitle,
			LoadDescription:        load.LoadDescription,
			LoadImage:              load.LoadImage,
			Normative:              load.Normative,
			LoadCategory:           load.LoadCategory,
			ReliabilityCoefficient: load.ReliabilityCoefficient,
			Status:                 load.Status,
		})
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: loadDTOs,
		Total: int64(len(loadDTOs)),
	})
}

func (h *Handler) GetLoadByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	load, err := h.Repository.GetLoadByID(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	loadDTO := ds.LoadDTO{
		ID:                     load.ID,
		LoadTitle:              load.LoadTitle,
		LoadDescription:        load.LoadDescription,
		LoadImage:              load.LoadImage,
		Normative:              load.Normative,
		LoadCategory:           load.LoadCategory,
		ReliabilityCoefficient: load.ReliabilityCoefficient,
		Status:                 load.Status,
	}

	c.JSON(http.StatusOK, loadDTO)
}

func (h *Handler) CreateLoad(c *gin.Context) {
	var req ds.LoadCreateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	statusValue := false
	load := ds.Loads{
		LoadTitle:              req.LoadTitle,
		LoadDescription:        req.LoadDescription,
		Normative:              req.Normative,
		LoadCategory:           req.LoadCategory,
		ReliabilityCoefficient: req.ReliabilityCoefficient,
		Status:                 &statusValue,
	}

	if err := h.Repository.CreateLoad(&load); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	loadDTO := ds.LoadDTO{
		ID:                     load.ID,
		LoadTitle:              load.LoadTitle,
		LoadDescription:        load.LoadDescription,
		LoadImage:              load.LoadImage,
		Normative:              load.Normative,
		LoadCategory:           load.LoadCategory,
		ReliabilityCoefficient: load.ReliabilityCoefficient,
		Status:                 load.Status,
	}

	c.JSON(http.StatusCreated, loadDTO)
}

func (h *Handler) UpdateLoad(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.LoadUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	load, err := h.Repository.UpdateLoad(uint(id), req)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	loadDTO := ds.LoadDTO{
		ID:                     load.ID,
		LoadTitle:              load.LoadTitle,
		LoadDescription:        load.LoadDescription,
		LoadImage:              load.LoadImage,
		Normative:              load.Normative,
		LoadCategory:           load.LoadCategory,
		ReliabilityCoefficient: load.ReliabilityCoefficient,
		Status:                 load.Status,
	}

	c.JSON(http.StatusOK, loadDTO)
}

func (h *Handler) DeleteLoad(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteLoad(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, ds.SuccessResponse{
		Message: "Нагрузка удалена",
	})
}

func (h *Handler) UploadLoadImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadLoadImage(uint(id), file)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": imageURL})
}

func (h *Handler) AddLoadToDraft(c *gin.Context) {
	loadID, err := strconv.Atoi(c.Param("load_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.AddLoadToDraft(hardcodedUserID, uint(loadID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, ds.SuccessResponse{
		Message: "Черновик создан. Нагрузка добавлена в черновик.",
	})
}
