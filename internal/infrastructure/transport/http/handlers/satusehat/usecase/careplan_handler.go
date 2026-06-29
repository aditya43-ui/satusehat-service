package usecase

import (
	"net/http"
	"service/internal/satusehat/usecase/careplan"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type CarePlanHandler struct{ service careplan.Service }

func NewCarePlanHandler(s careplan.Service) *CarePlanHandler { return &CarePlanHandler{service: s} }

func (h *CarePlanHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/careplan")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *CarePlanHandler) Create(c *gin.Context) {
	var req careplan.CarePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Success", res.FullResponse)
}
func (h *CarePlanHandler) Update(c *gin.Context) {
	var req careplan.CarePlanRequest
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
func (h *CarePlanHandler) Patch(c *gin.Context) {
	var req careplan.CarePlanPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", validator.TranslateError(err))
		return
	}
	res, err := h.service.Patch(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *CarePlanHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *CarePlanHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
