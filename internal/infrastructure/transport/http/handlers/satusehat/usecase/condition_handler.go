package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/condition"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ConditionHandler struct {
	service condition.Service
}

func NewConditionHandler(service condition.Service) *ConditionHandler {
	return &ConditionHandler{
		service: service,
	}
}

func (h *ConditionHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/condition")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *ConditionHandler) Create(c *gin.Context) {
	var req condition.ConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", customErr)
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created Condition", result)
}

func (h *ConditionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req condition.ConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", customErr)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated Condition", result)
}

func (h *ConditionHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req condition.ConditionPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", customErr)
		return
	}

	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched Condition", result)
}

func (h *ConditionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Condition", result)
}

func (h *ConditionHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Conditions", result)
}
