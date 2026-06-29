package usecase

import (
	"net/http"
	"service/internal/satusehat/usecase/immunization"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ImmunizationHandler struct{ service immunization.Service }

func NewImmunizationHandler(s immunization.Service) *ImmunizationHandler {
	return &ImmunizationHandler{service: s}
}

func (h *ImmunizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/immunization")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *ImmunizationHandler) Create(c *gin.Context) {
	var req immunization.ImmunizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Invalid request", validator.TranslateError(err))
		return
	}
	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Success", res.FullResponse)
}
func (h *ImmunizationHandler) Update(c *gin.Context) {
	var req immunization.ImmunizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Invalid request", validator.TranslateError(err))
		return
	}
	res, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *ImmunizationHandler) Patch(c *gin.Context) {
	var req immunization.ImmunizationPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Invalid patch", validator.TranslateError(err))
		return
	}
	res, err := h.service.Patch(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *ImmunizationHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *ImmunizationHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
