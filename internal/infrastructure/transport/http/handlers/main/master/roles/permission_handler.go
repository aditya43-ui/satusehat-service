package role

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	roleService "service/internal/master/role/permission"
	"service/pkg/errors"
	"service/pkg/logger"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type RolPermissionHandler struct {
	service roleService.Service
}

func NewRolPermissionHandler(service roleService.Service) *RolPermissionHandler {
	return &RolPermissionHandler{service: service}
}

func (h *RolPermissionHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/roles/permissions") //  get by role
	{
		group.GET("", h.GetList)
		group.GET("/search", h.Search)
		group.GET("/:id", h.GetDetail)                    // using role id
		group.GET("/role/:role", h.GetPermissionTreeRole) // get permission tree by role keycloak
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
		group.GET("/rolemaster/:id", h.GetRolePagesList)
	}
}

// GetList godoc
//
//	@Summary		Get list of Role Permissions
//	@Description	Retrieve a paginated list of Role Permissions
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			limit	query		int		false	"Number of items per page"	default(10)
//	@Param			offset	query		int		false	"Offset for pagination"		default(0)
//	@Param			sort	query		string	false	"Sort fields (e.g. +name,-created_at)"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions [get]
func (h *RolPermissionHandler) GetList(c *gin.Context) {
	// Parse Limit & Offset dengan fallback ke page & page_size
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if pageSizeStr := c.Query("limit"); pageSizeStr != "" {
		limit, _ = strconv.Atoi(pageSizeStr)
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if pageStr := c.Query("offset"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 && limit > 0 {
			offset = (page - 1) * limit
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
	response.Paginated(c, http.StatusOK, "Successfully retrieved RolPermission list", data, meta)
}

// GetDetail godoc
//
//	@Summary		Get Role Permission detail
//	@Description	Retrieve detailed information about a specific Role Permission
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			id	path		int	true	"Role Permission ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/{id} [get]
func (h *RolPermissionHandler) GetDetail(c *gin.Context) {
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

	response.Success(c, http.StatusOK, "Successfully retrieved RolPermission detail", result)
}

// Search godoc
//
//	@Summary		Search Role Permissions
//	@Description	Search Role Permissions records using dynamic filters
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			limit	query		int		false	"Limit per page"		default(10)
//	@Param			offset	query		int		false	"Offset for pagination"	default(0)
//	@Param			sort	query		string	false	"Sort fields"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/search [get]
func (h *RolPermissionHandler) Search(c *gin.Context) {
	// Parse Limit & Offset dengan fallback ke page & page_size
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if pageSizeStr := c.Query("limit"); pageSizeStr != "" {
		limit, _ = strconv.Atoi(pageSizeStr)
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if pageStr := c.Query("offset"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 && limit > 0 {
			offset = (page - 1) * limit
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

	response.Paginated(c, http.StatusOK, "Successfully retrieved RolPermission search results", data, meta)
}

// GetPermissionTree godoc
//
//	@Summary		Get Permission Tree
//	@Description	Retrieve a permission tree
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			role	path		string	true	"Role Keycloak"
//	@Param			groups	query		string	false	"Comma-separated groups"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/tree/{role} [get]
func (h *RolPermissionHandler) GetPermissionTree(c *gin.Context) {
	roleKeycloak := c.Param("role")

	// Get groups from query parameter (comma-separated)
	groupsParam := c.Query("groups")
	var groups []string
	if groupsParam != "" {
		groups = strings.Split(groupsParam, ",")
		// Trim spaces from each group
		for i, group := range groups {
			groups[i] = strings.TrimSpace(group)
		}
	}

	// Parse active parameter - handle "true", "false", "1", "0"
	activeParam := c.Query("active")
	var activeOnly *bool // default nil (no filter)
	if activeParam != "" {
		isActive := activeParam == "true" || activeParam == "1"
		activeOnly = &isActive
	}

	ctx := c.Request.Context()
	result, err := h.service.GetRolePermissionTree(ctx, roleKeycloak, groups, activeOnly)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	// Jika role disediakan tapi tidak ada data, return empty result
	if roleKeycloak != "" && (result == nil || len(result.Data.Access) == 0) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "No permissions found for the specified role",
			"data":    []interface{}{},
		})
		return
	}

	// Return the result in the desired format
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully retrieved RolPages",
		"data":    result.Data,
	})
}

// GetPermissionTreeRole godoc
//
//	@Summary		Get Permission Tree by Role
//	@Description	Retrieve a permission tree by Keycloak role
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			role	path		string	true	"Role Keycloak"
//	@Param			active	query		boolean	false	"Filter by active status"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/role/{role} [get]
func (h *RolPermissionHandler) GetPermissionTreeRole(c *gin.Context) {
	roleKeycloak := c.Param("role")

	// Parse active parameter - handle "true", "false", "1", "0"
	activeParam := c.Query("active")
	var activeOnly *bool // default nil (no filter)
	if activeParam != "" {
		isActive := activeParam == "true" || activeParam == "1"
		activeOnly = &isActive
	}

	ctx := c.Request.Context()
	result, err := h.service.GetRolePermissionTreeRole(ctx, roleKeycloak, activeOnly)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	// Jika role disediakan tapi tidak ada data, return empty result
	if roleKeycloak != "" && (result == nil || len(result.Data.Access) == 0) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "No permissions found for the specified role",
			"data":    []interface{}{},
		})
		return
	}

	// Return the result in the desired format
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully retrieved RolPages",
		"data":    result.Data,
	})
}

// Create godoc
//
//	@Summary		Create new Role Permission
//	@Description	Create a new Role Permission record
//	@Tags			roles-permissions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		roleService.RolPermissionRequest	true	"Payload"
//	@Success		201		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions [post]
func (h *RolPermissionHandler) Create(c *gin.Context) {
	var req roleService.RolPermissionRequest
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

	response.Success(c, http.StatusCreated, "Successfully created RolPermission", created)
}

// Update godoc
//
//	@Summary		Update an existing Role Permission
//	@Description	Update details of an existing Role Permission record by ID
//	@Tags			roles-permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int									true	"Role Permission ID"
//	@Param			request	body		roleService.RolPermissionRequest	true	"Payload"
//	@Success		200		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/{id} [put]
func (h *RolPermissionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req roleService.RolPermissionRequest
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

	response.Success(c, http.StatusOK, "Successfully updated RolPermission", updated)
}

// Delete godoc
//
//	@Summary		Delete a Role Permission
//	@Description	Delete a Role Permission record by ID
//	@Tags			roles-permissions
//	@Produce		json
//	@Param			id	path		int	true	"Role Permission ID"
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/roles/permissions/{id} [delete]
func (h *RolPermissionHandler) Delete(c *gin.Context) {
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

	response.Success(c, http.StatusOK, "Successfully deleted RolPermission", nil)
}

// GetRolePagesList mengambil daftar role beserta page dan permission-nya
// @Summary Get Role Pages Access List
// @Description Retrieve a paginated list of RolePages grouped by Role with Tree Hierarchy
// @Tags roles-permissions
// @Produce json
// @Param id path string true "Role Master ID"
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Items per page" default(10)
// @Param active query boolean false "Filter by active status"
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /roles/permissions/rolemaster/{id} [get]
func (h *RolPermissionHandler) GetRolePagesList(c *gin.Context) {
	ctx := c.Request.Context()
	roleID := c.Param("id")

	// 1. Ambil Parameter Pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < -1 {
		limit = 10
	}

	offset := (page - 1) * limit
	if limit == -1 {
		offset = 0 // Jika limit -1, fetch all data
	}

	// 2. Ambil Parameter Filter Active
	defaultActive := true
	activeFilter := &defaultActive
	if activeStr := c.Query("active"); activeStr != "" {
		parsedActive, err := strconv.ParseBool(activeStr)
		if err == nil {
			activeFilter = &parsedActive
		}
	}

	// 3. Panggil Service layer
	res, err := h.service.GetRolePagesList(ctx, roleID, limit, offset, activeFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 4. Kalkulasi Metadata & Response Sesuai Format
	total := res["total"].(int64)
	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	// Format JSON Response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully retrieved RolPages list",
		"data":    res["data"],
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
