package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/imagingstudy"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ImagingStudyHandler struct {
	service imagingstudy.Service
}

func NewImagingStudyHandler(service imagingstudy.Service) *ImagingStudyHandler {
	return &ImagingStudyHandler{
		service: service,
	}
}

func (h *ImagingStudyHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/imagingstudy")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *ImagingStudyHandler) Create(c *gin.Context) {
	var req imagingstudy.ImagingStudyRequest
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
	response.Success(c, http.StatusCreated, "Successfully created ImagingStudy", result)
}

func (h *ImagingStudyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req imagingstudy.ImagingStudyRequest
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
	response.Success(c, http.StatusOK, "Successfully updated ImagingStudy", result)
}

func (h *ImagingStudyHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req imagingstudy.ImagingStudyPatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched ImagingStudy", result)
}

func (h *ImagingStudyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ImagingStudy", result)
}

func (h *ImagingStudyHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ImagingStudies", result)
}
