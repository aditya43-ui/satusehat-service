package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/medicationrequest"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type MedicationRequestHandler struct {
	service medicationrequest.Service
}

func NewMedicationRequestHandler(service medicationrequest.Service) *MedicationRequestHandler {
	return &MedicationRequestHandler{service: service}
}

func (h *MedicationRequestHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/medicationrequest")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *MedicationRequestHandler) Create(c *gin.Context) {
	var req medicationrequest.MedicationRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created MedicationRequest", result.FullResponse)
}

func (h *MedicationRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req medicationrequest.MedicationRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated MedicationRequest", result.FullResponse)
}

func (h *MedicationRequestHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req medicationrequest.MedicationRequestPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched MedicationRequest", result.FullResponse)
}

func (h *MedicationRequestHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
func (h *MedicationRequestHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
