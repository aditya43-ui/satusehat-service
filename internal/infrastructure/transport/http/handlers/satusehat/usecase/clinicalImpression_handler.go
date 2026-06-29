package usecase

import (
	"net/http"
	clinicalimpression "service/internal/satusehat/usecase/clinicalImpression"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ClinicalImpressionHandler struct{ service clinicalimpression.Service }

func NewClinicalImpressionHandler(s clinicalimpression.Service) *ClinicalImpressionHandler {
	return &ClinicalImpressionHandler{service: s}
}

func (h *ClinicalImpressionHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/clinicalimpression")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *ClinicalImpressionHandler) Create(c *gin.Context) {
	var req clinicalimpression.ClinicalImpressionRequest
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
func (h *ClinicalImpressionHandler) Update(c *gin.Context) {
	var req clinicalimpression.ClinicalImpressionRequest
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
func (h *ClinicalImpressionHandler) Patch(c *gin.Context) {
	var req clinicalimpression.ClinicalImpressionPatchRequest
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
func (h *ClinicalImpressionHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *ClinicalImpressionHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
