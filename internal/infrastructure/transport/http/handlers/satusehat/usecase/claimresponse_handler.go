package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/claimresponse"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ClaimResponseHandler struct {
	service claimresponse.Service
}

func NewClaimResponseHandler(service claimresponse.Service) *ClaimResponseHandler {
	return &ClaimResponseHandler{service: service}
}

func (h *ClaimResponseHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/claim-response")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
	}
}

func (h *ClaimResponseHandler) Create(c *gin.Context) {
	var req claimresponse.ClaimResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Successfully created ClaimResponse", result.FullResponse)
}

func (h *ClaimResponseHandler) Update(c *gin.Context) {
	var req claimresponse.ClaimResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", validator.TranslateError(err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated ClaimResponse", result.FullResponse)
}

func (h *ClaimResponseHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ClaimResponse", result.FullResponse)
}

func (h *ClaimResponseHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved ClaimResponses", result.FullResponse)
}
