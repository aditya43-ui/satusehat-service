package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/medicationdispense"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type MedicationDispenseHandler struct {
	service medicationdispense.Service
}

func NewMedicationDispenseHandler(service medicationdispense.Service) *MedicationDispenseHandler {
	return &MedicationDispenseHandler{service: service}
}

func (h *MedicationDispenseHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/medicationdispense")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *MedicationDispenseHandler) Create(c *gin.Context) {
	var req medicationdispense.MedicationDispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created MedicationDispense", result.FullResponse)
}

func (h *MedicationDispenseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req medicationdispense.MedicationDispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated MedicationDispense", result.FullResponse)
}

func (h *MedicationDispenseHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req medicationdispense.MedicationDispensePatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched MedicationDispense", result.FullResponse)
}

func (h *MedicationDispenseHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
func (h *MedicationDispenseHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
