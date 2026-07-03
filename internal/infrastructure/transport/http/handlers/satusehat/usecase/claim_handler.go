package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/claim"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ClaimHandler struct {
	service claim.Service
}

func NewClaimHandler(service claim.Service) *ClaimHandler {
	return &ClaimHandler{service: service}
}

func (h *ClaimHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/claim")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
	}
}

func (h *ClaimHandler) Create(c *gin.Context) {
	var req claim.ClaimRequest
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
	response.Success(c, http.StatusCreated, "Successfully created Claim", result.FullResponse)
}

func (h *ClaimHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req claim.ClaimRequest
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
	response.Success(c, http.StatusOK, "Successfully updated Claim", result.FullResponse)
}

func (h *ClaimHandler) GetByID(c *gin.Context) {
	result, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Claim", result.FullResponse)
}

func (h *ClaimHandler) Search(c *gin.Context) {
	result, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Claims", result.FullResponse)
}
