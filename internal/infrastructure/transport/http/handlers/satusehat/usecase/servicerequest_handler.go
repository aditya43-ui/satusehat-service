package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/servicerequest"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ServiceRequestHandler struct {
	service servicerequest.Service
}

func NewServiceRequestHandler(service servicerequest.Service) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		service: service,
	}
}

func (h *ServiceRequestHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/servicerequest")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *ServiceRequestHandler) Create(c *gin.Context) {
	var req servicerequest.ServiceRequestRequest
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
	response.Success(c, http.StatusCreated, "Successfully created ServiceRequest", result.FullResponse)
}

func (h *ServiceRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req servicerequest.ServiceRequestRequest
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
	response.Success(c, http.StatusOK, "Successfully updated ServiceRequest", result.FullResponse)
}

func (h *ServiceRequestHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req servicerequest.ServiceRequestPatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched ServiceRequest", result.FullResponse)
}

func (h *ServiceRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ServiceRequest", result.FullResponse)
}

func (h *ServiceRequestHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ServiceRequests", result.FullResponse)
}
