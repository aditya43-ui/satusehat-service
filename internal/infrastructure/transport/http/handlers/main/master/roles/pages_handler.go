package role

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	roleService "service/internal/master/role/pages"
	"service/pkg/errors"
	"service/pkg/logger"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type RolPagesHandler struct {
	service roleService.Service
}

func NewRolPagesHandler(service roleService.Service) *RolPagesHandler {
	return &RolPagesHandler{service: service}
}

func (h *RolPagesHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/roles/pages")
	{
		group.GET("", h.GetList)
		group.GET("/tree", h.GetTree)
		group.GET("/tree/level/:level", h.GetTreeByLevel)
		group.GET("/search", h.Search)
		group.GET("/:id", h.GetDetail)
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

// GetList godoc
//
//	@Summary		Get list of Role Pages
//	@Description	Retrieve a paginated list of Role Pages
//	@Tags			roles-pages
//	@Produce		json
//	@Param			limit	query		int		false	"Number of items per page"	default(10)
//	@Param			offset	query		int		false	"Offset for pagination"		default(0)
//	@Param			sort	query		string	false	"Sort fields (e.g. +name,-created_at)"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages [get]
func (h *RolPagesHandler) GetList(c *gin.Context) {
	// Parse Limit & Offset dengan fallback ke page & page_size
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset > 0 && limit > 0 {
			offset = (offset - 1) * limit
		}
	}

	activeParam := c.Query("active")
	var activeFilter *bool // default nil (no filter / ambil semua)
	if activeParam != "" {
		isActive := activeParam == "true" || activeParam == "1"
		activeFilter = &isActive
	}

	// Parse parameter sort (format: sort=column1,-column2,+column3)
	// -column untuk DESC, +column atau column untuk ASC
	var sorts []string
	if sortParam := c.Query("sort"); sortParam != "" {
		sorts = strings.Split(sortParam, ",")
		// Validasi dan bersihkan sort parameters
		for i, sort := range sorts {
			sorts[i] = strings.TrimSpace(sort)
		}
	}

	ctx := c.Request.Context()
	result, err := h.service.GetList(ctx, limit, offset, sorts, activeFilter)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	data := result["data"]
	total := result["total"].(int64)
	limitVal := result["limit"].(int)
	offsetVal := result["offset"].(int)

	page := 1
	if limitVal > 0 {
		page = (offsetVal / limitVal) + 1
	}
	totalPages := 0
	if limitVal > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limitVal)))
	}

	meta := response.Meta{Page: page, Limit: limitVal, Total: int(total), TotalPages: totalPages}
	response.Paginated(c, http.StatusOK, "Successfully retrieved RolPages list", data, meta)
}

// GetDetail godoc
//
//	@Summary		Get Role Pages detail
//	@Description	Retrieve detailed information about a specific Role Pages
//	@Tags			roles-pages
//	@Produce		json
//	@Param			id	path		int	true	"Role Pages ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/{id} [get]
func (h *RolPagesHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	ctx := c.Request.Context()
	result, err := h.service.GetDetail(ctx, id)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Successfully retrieved RolPages detail", result)
}

// Search godoc
//
//	@Summary		Search Role Pages
//	@Description	Search Role Pages records using dynamic filters
//	@Tags			roles-pages
//	@Produce		json
//	@Param			limit	query		int		false	"Limit per page"		default(10)
//	@Param			offset	query		int		false	"Offset for pagination"	default(0)
//	@Param			sort	query		string	false	"Sort fields"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/search [get]
func (h *RolPagesHandler) Search(c *gin.Context) {
	// Parse Limit & Offset dengan fallback ke page & page_size
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset > 0 && limit > 0 {
			offset = (offset - 1) * limit
		}
	}

	// Ambil parameter filter secara dinamis
	filters := make(map[string]interface{})

	for key, values := range c.Request.URL.Query() {
		if key == "page" || key == "limit" || key == "page_size" || key == "offset" || key == "sort" {
			continue
		}
		if len(values) > 0 && values[0] != "" {
			filters[key] = values[0]
		}
	}

	// Parse parameter sort (format: sort=column1,-column2,+column3)
	// -column untuk DESC, +column atau column untuk ASC
	var sorts []string
	if sortParam := c.Query("sort"); sortParam != "" {
		sorts = strings.Split(sortParam, ",")
		// Validasi dan bersihkan sort parameters
		for i, sort := range sorts {
			sorts[i] = strings.TrimSpace(sort)
		}
	}

	ctx := c.Request.Context()

	logger.Default().WithContext(ctx).Info("Search request",
		logger.String("filters", fmt.Sprintf("%v", filters)),
		logger.String("sorts", fmt.Sprintf("%v", sorts)),
		logger.Int("offset", offset),
		logger.Int("limit", limit))

	// Panggil service dengan parameter sort tambahan
	result, err := h.service.Search(ctx, filters, sorts, limit, offset)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	// Extract data dari map service untuk response format
	data := result["data"]
	total := result["total"].(int64)
	limitVal := result["limit"].(int)
	offsetVal := result["offset"].(int)

	page := 1
	if limitVal > 0 {
		page = (offsetVal / limitVal) + 1
	}
	totalPages := 0
	if limitVal > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limitVal)))
	}

	meta := response.Meta{
		Page:       page,
		Limit:      limitVal,
		Total:      int(total),
		TotalPages: totalPages,
	}

	response.Paginated(c, http.StatusOK, "Successfully retrieved RolPages search results", data, meta)
}

// Create godoc
//
//	@Summary		Create new Role Pages
//	@Description	Create a new Role Pages record
//	@Tags			roles-pages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		roleService.RolPagesRequest	true	"Payload"
//	@Success		201		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages [post]
func (h *RolPagesHandler) Create(c *gin.Context) {
	var req roleService.RolPagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx := c.Request.Context()
	created, err := h.service.Create(ctx, req)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusCreated, "Successfully created RolPages", created)
}

// Update godoc
//
//	@Summary		Update an existing Role Pages
//	@Description	Update details of an existing Role Pages record by ID
//	@Tags			roles-pages
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Role Pages ID"
//	@Param			request	body		roleService.RolPagesRequest	true	"Payload"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/{id} [put]
func (h *RolPagesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req roleService.RolPagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx := c.Request.Context()
	updated, err := h.service.Update(ctx, id, req)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Successfully updated RolPages", updated)
}

// Delete godoc
//
//	@Summary		Delete a Role Pages
//	@Description	Delete a Role Pages record by ID
//	@Tags			roles-pages
//	@Produce		json
//	@Param			id	path		int	true	"Role Pages ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/{id} [delete]
func (h *RolPagesHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Delete(ctx, id); err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Successfully deleted RolPages", nil)
}

// GetTree godoc
//
//	@Summary		Get tree of Role Pages
//	@Description	Retrieve a hierarchical tree of Role Pages
//	@Tags			roles-pages
//	@Produce		json
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/tree [get]
func (h *RolPagesHandler) GetTree(c *gin.Context) {
	activeParam := c.Query("active")
	var activeFilter *bool // default nil (no filter / ambil semua)
	if activeParam != "" {
		isActive := activeParam == "true" || activeParam == "1"
		activeFilter = &isActive
	}

	ctx := c.Request.Context()
	tree, err := h.service.GetTree(ctx, activeFilter)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Successfully retrieved RolPages", tree)
}

// GetTreeByLevel godoc
//
//	@Summary		Get tree of Role Pages by level
//	@Description	Retrieve a hierarchical tree of Role Pages up to a specific level
//	@Tags			roles-pages
//	@Produce		json
//	@Param			level	path		int		true	"Hierarchy Level"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/pages/tree/level/{level} [get]
func (h *RolPagesHandler) GetTreeByLevel(c *gin.Context) {
	level, err := strconv.ParseInt(c.Param("level"), 10, 16)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid level format", nil)
		return
	}

	activeParam := c.Query("active")
	var activeFilter *bool // default nil (no filter / ambil semua)
	if activeParam != "" {
		isActive := activeParam == "true" || activeParam == "1"
		activeFilter = &isActive
	}

	ctx := c.Request.Context()
	tree, err := h.service.GetTreeByLevel(ctx, int16(level), activeFilter)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Successfully retrieved RolPages", tree)
}
