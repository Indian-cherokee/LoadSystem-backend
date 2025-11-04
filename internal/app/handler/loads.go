package handler

import (
	"net/http"
	"strconv"
	"web/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// GetLoads godoc
// @Summary      Получить список нагрузок (все)
// @Description  Возвращает отфильтрованный список всех нагрузок. Доступно всем.
// @Tags         loads
// @Produce      json
// @Param        search query string false "Поиск по названию"
// @Param        category query string false "Фильтр по категории"
// @Success      200 {object} ds.PaginatedResponse
// @Router       /loads [get]
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

// GetLoadByID godoc
// @Summary      Получить нагрузку по ID (все)
// @Description  Возвращает информацию о нагрузке по её ID. Доступно всем.
// @Tags         loads
// @Produce      json
// @Param        id path int true "ID нагрузки"
// @Success      200 {object} ds.LoadDTO
// @Failure      404 {object} map[string]string "Нагрузка не найдена"
// @Router       /loads/{id} [get]
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

// CreateLoad godoc
// @Summary      Создать новую нагрузку (модератор)
// @Description  Создает новую нагрузку в системе. Требует прав модератора.
// @Tags         loads
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        load body ds.LoadCreateRequest true "Данные нагрузки"
// @Success      201 {object} ds.LoadDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Требуются права модератора"
// @Router       /loads [post]
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

// UpdateLoad godoc
// @Summary      Обновить нагрузку (модератор)
// @Description  Обновляет данные нагрузки. Требует прав модератора.
// @Tags         loads
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID нагрузки"
// @Param        load body ds.LoadUpdateRequest true "Данные для обновления"
// @Success      200 {object} ds.LoadDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Требуются права модератора"
// @Router       /loads/{id} [put]
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

// DeleteLoad godoc
// @Summary      Удалить нагрузку (модератор)
// @Description  Удаляет нагрузку из системы. Требует прав модератора.
// @Tags         loads
// @Security     ApiKeyAuth
// @Param        id path int true "ID нагрузки"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Требуются права модератора"
// @Router       /loads/{id} [delete]
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

// UploadLoadImage godoc
// @Summary      Загрузить изображение для нагрузки (модератор)
// @Description  Загружает изображение для нагрузки. Требует прав модератора.
// @Tags         loads
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID нагрузки"
// @Param        file formData file true "Изображение"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Требуются права модератора"
// @Router       /loads/{id}/image [post]
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

// AddLoadToDraft godoc
// @Summary      Добавить нагрузку в черновик (авторизованный пользователь)
// @Description  Добавляет нагрузку в черновик сессии текущего пользователя. Если черновика нет, создает новый.
// @Tags         load-sessions
// @Security     ApiKeyAuth
// @Param        load_id path int true "ID нагрузки"
// @Success      201 {object} ds.SuccessResponse
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /load-sessions/draft/loads/{load_id} [post]
func (h *Handler) AddLoadToDraft(c *gin.Context) {
	loadID, err := strconv.Atoi(c.Param("load_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.AddLoadToDraft(userID, uint(loadID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, ds.SuccessResponse{
		Message: "Черновик создан. Нагрузка добавлена в черновик.",
	})
}
