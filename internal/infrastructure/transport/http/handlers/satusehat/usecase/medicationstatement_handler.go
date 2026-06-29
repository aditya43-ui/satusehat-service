package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/medicationstatement"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type MedicationStatementHandler struct {
	service medicationstatement.Service
}

func NewMedicationStatementHandler(service medicationstatement.Service) *MedicationStatementHandler {
	return &MedicationStatementHandler{service: service}
}

func (h *MedicationStatementHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/medicationstatement")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *MedicationStatementHandler) Create(c *gin.Context) {
	var req medicationstatement.MedicationStatementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created MedicationStatement", result.FullResponse)
}

func (h *MedicationStatementHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req medicationstatement.MedicationStatementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated MedicationStatement", result.FullResponse)
}

func (h *MedicationStatementHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req medicationstatement.MedicationStatementPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched MedicationStatement", result.FullResponse)
}

func (h *MedicationStatementHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
func (h *MedicationStatementHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
