package usecase

import (
	"net/http"
	"service/internal/satusehat/usecase/specimen"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type SpecimenHandler struct{ service specimen.Service }

func NewSpecimenHandler(s specimen.Service) *SpecimenHandler { return &SpecimenHandler{service: s} }

func (h *SpecimenHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/specimen")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *SpecimenHandler) Create(c *gin.Context) {
	var req specimen.SpecimenRequest
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
func (h *SpecimenHandler) Update(c *gin.Context) {
	var req specimen.SpecimenRequest
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
func (h *SpecimenHandler) Patch(c *gin.Context) {
	var req specimen.SpecimenPatchRequest
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
func (h *SpecimenHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *SpecimenHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
