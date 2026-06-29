package role

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	masterService "service/internal/master/role/master"
	"service/pkg/errors"
	"service/pkg/logger"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleMasterHandler struct {
	service masterService.Service
}

func NewRoleMasterHandler(service masterService.Service) *RoleMasterHandler {
	return &RoleMasterHandler{service: service}
}

func (h *RoleMasterHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/roles/master")
	{
		group.GET("", h.GetList)
		group.GET("/search", h.Search)
		group.GET("/:id", h.GetDetail)
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

// GetList godoc
//
//	@Summary		Get list of Role
//	@Description	Retrieve a paginated list of RoleMaster
//	@Tags			roles
//	@Produce		json
//	@Param			limit	query		int		false	"Number of items per page"	default(10)
//	@Param			offset	query		int		false	"Offset for pagination"		default(0)
//	@Param			sort	query		string	false	"Sort fields (e.g. +name,-created_at)"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master [get]
func (h *RoleMasterHandler) GetList(c *gin.Context) {
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
//	@Summary		Get RoleMaster detail
//	@Description	Retrieve detailed information about a specific RoleMaster
//	@Tags			roles
//	@Produce		json
//	@Param			id	path		int	true	"RoleMaster ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master/{id} [get]
func (h *RoleMasterHandler) GetDetail(c *gin.Context) {
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

	response.Success(c, http.StatusOK, "Successfully retrieved RoleMaster detail", result)
}

// Search godoc
//
//	@Summary		Search Role
//	@Description	Search RoleMaster records using dynamic filters
//	@Tags			roles
//	@Produce		json
//	@Param			limit	query		int		false	"Limit per page"		default(10)
//	@Param			offset	query		int		false	"Offset for pagination"	default(0)
//	@Param			sort	query		string	false	"Sort fields (e.g. +name,-created_at)"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master/search [get]
func (h *RoleMasterHandler) Search(c *gin.Context) {
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

	// Ambil parameter filter
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
	// Hitung total pages
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

	response.Paginated(c, http.StatusOK, "Successfully retrieved RoleMaster search results", data, meta)
}

// Create godoc
//
//	@Summary		Create new RoleMaster
//	@Description	Create a new RoleMaster record
//	@Tags			roles
//	@Accept			json
//	@Produce		json
//	@Param			request	body		masterService.RoleMasterRequest	true	"Payload"
//	@Success		201		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master [post]
func (h *RoleMasterHandler) Create(c *gin.Context) {
	var req masterService.RoleMasterRequest
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

	response.Success(c, http.StatusCreated, "Successfully created RoleMaster", created)
}

// Update godoc
//
//	@Summary		Update an existing RoleMaster
//	@Description	Update details of an existing RoleMaster record by ID
//	@Tags			roles
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"RoleMaster ID"
//	@Param			request	body		masterService.RoleMasterRequest	true	"Payload"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master/{id} [put]
func (h *RoleMasterHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req masterService.RoleMasterRequest
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

	response.Success(c, http.StatusOK, "Successfully updated RoleMaster", updated)
}

// Delete godoc
//
//	@Summary		Delete a RoleMaster
//	@Description	Delete a RoleMaster record by ID (soft delete)
//	@Tags			roles
//	@Produce		json
//	@Param			id	path		int	true	"RoleMaster ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/master/{id} [delete]
func (h *RoleMasterHandler) Delete(c *gin.Context) {
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

	response.Success(c, http.StatusOK, "Successfully deleted RoleMaster", nil)
}
