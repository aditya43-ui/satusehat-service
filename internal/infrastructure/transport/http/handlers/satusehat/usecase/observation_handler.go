package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/observation"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ObservationHandler struct {
	service observation.Service
}

func NewObservationHandler(service observation.Service) *ObservationHandler {
	return &ObservationHandler{
		service: service,
	}
}

func (h *ObservationHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/observation")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *ObservationHandler) Create(c *gin.Context) {
	var req observation.ObservationRequest
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
	response.Success(c, http.StatusCreated, "Successfully created Observation", result.FullResponse)
}

func (h *ObservationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req observation.ObservationRequest
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
	response.Success(c, http.StatusOK, "Successfully updated Observation", result.FullResponse)
}

func (h *ObservationHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req observation.ObservationPatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched Observation", result.FullResponse)
}

func (h *ObservationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Observation", result.FullResponse)
}

func (h *ObservationHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Observations", result.FullResponse)
}
