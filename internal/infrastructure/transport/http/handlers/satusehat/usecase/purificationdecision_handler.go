package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/purificationdecision"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type PurificationDecisionHandler struct {
	service purificationdecision.Service
}

func NewPurificationDecisionHandler(service purificationdecision.Service) *PurificationDecisionHandler {
	return &PurificationDecisionHandler{service: service}
}

func (h *PurificationDecisionHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/purification-decision")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
	}
}

func (h *PurificationDecisionHandler) Create(c *gin.Context) {
	var req purificationdecision.PurificationDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created PurificationDecision", result.FullResponse)
}

func (h *PurificationDecisionHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved PurificationDecision", result.FullResponse)
}

func (h *PurificationDecisionHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved PurificationDecisions", result.FullResponse)
}
