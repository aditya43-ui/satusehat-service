package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/questionnaireresponse"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type QuestionnaireResponseHandler struct {
	service questionnaireresponse.Service
}

func NewQuestionnaireResponseHandler(service questionnaireresponse.Service) *QuestionnaireResponseHandler {
	return &QuestionnaireResponseHandler{service: service}
}

func (h *QuestionnaireResponseHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/questionnaireresponse")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *QuestionnaireResponseHandler) Create(c *gin.Context) {
	var req questionnaireresponse.QuestionnaireResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created QuestionnaireResponse", result.FullResponse)
}

func (h *QuestionnaireResponseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req questionnaireresponse.QuestionnaireResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated QuestionnaireResponse", result.FullResponse)
}

func (h *QuestionnaireResponseHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req questionnaireresponse.QuestionnaireResponsePatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched QuestionnaireResponse", result.FullResponse)
}

func (h *QuestionnaireResponseHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
func (h *QuestionnaireResponseHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved", result.FullResponse)
}
