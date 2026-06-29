package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/composition"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type CompositionHandler struct {
	service composition.Service
}

func NewCompositionHandler(service composition.Service) *CompositionHandler {
	return &CompositionHandler{
		service: service,
	}
}

func (h *CompositionHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/composition")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *CompositionHandler) Create(c *gin.Context) {
	var req composition.CompositionRequest
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
	response.Success(c, http.StatusCreated, "Successfully created Composition", result.FullResponse)
}

func (h *CompositionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req composition.CompositionRequest
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
	response.Success(c, http.StatusOK, "Successfully updated Composition", result.FullResponse)
}

func (h *CompositionHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req composition.CompositionPatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched Composition", result.FullResponse)
}

func (h *CompositionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Composition", result.FullResponse)
}

func (h *CompositionHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Compositions", result.FullResponse)
}
