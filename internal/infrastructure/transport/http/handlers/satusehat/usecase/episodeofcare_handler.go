package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/episodeofcare"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type EpisodeOfCareHandler struct {
	service episodeofcare.Service
}

func NewEpisodeOfCareHandler(service episodeofcare.Service) *EpisodeOfCareHandler {
	return &EpisodeOfCareHandler{
		service: service,
	}
}

func (h *EpisodeOfCareHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/episodeofcare")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *EpisodeOfCareHandler) Create(c *gin.Context) {
	var req episodeofcare.EpisodeOfCareRequest
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
	response.Success(c, http.StatusCreated, "Successfully created EpisodeOfCare", result)
}

func (h *EpisodeOfCareHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req episodeofcare.EpisodeOfCareRequest
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
	response.Success(c, http.StatusOK, "Successfully updated EpisodeOfCare", result)
}

func (h *EpisodeOfCareHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req episodeofcare.EpisodeOfCarePatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched EpisodeOfCare", result)
}

func (h *EpisodeOfCareHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved EpisodeOfCare", result)
}

func (h *EpisodeOfCareHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved EpisodeOfCare records", result)
}
