package usecase

import (
	"net/http"
	"service/internal/satusehat/usecase/allergyintolerance"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type AllergyIntoleranceHandler struct{ service allergyintolerance.Service }

func NewAllergyIntoleranceHandler(s allergyintolerance.Service) *AllergyIntoleranceHandler {
	return &AllergyIntoleranceHandler{service: s}
}

func (h *AllergyIntoleranceHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/allergyintolerance")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *AllergyIntoleranceHandler) Create(c *gin.Context) {
	var req allergyintolerance.AllergyIntoleranceRequest
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
func (h *AllergyIntoleranceHandler) Update(c *gin.Context) {
	var req allergyintolerance.AllergyIntoleranceRequest
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
func (h *AllergyIntoleranceHandler) Patch(c *gin.Context) {
	var req allergyintolerance.AllergyIntolerancePatchRequest
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
func (h *AllergyIntoleranceHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *AllergyIntoleranceHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
