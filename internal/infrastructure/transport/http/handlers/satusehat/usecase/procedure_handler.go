package usecase

import (
	"net/http"

	"service/internal/satusehat/usecase/procedure"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type ProcedureHandler struct {
	service procedure.Service
}

func NewProcedureHandler(service procedure.Service) *ProcedureHandler {
	return &ProcedureHandler{
		service: service,
	}
}

func (h *ProcedureHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/procedure")
	{
		group.POST("", h.Create)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

func (h *ProcedureHandler) Create(c *gin.Context) {
	var req procedure.ProcedureRequest
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
	response.Success(c, http.StatusCreated, "Successfully created Procedure", result.FullResponse)
}

func (h *ProcedureHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req procedure.ProcedureRequest
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
	response.Success(c, http.StatusOK, "Successfully updated Procedure", result.FullResponse)
}

func (h *ProcedureHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req procedure.ProcedurePatchRequest
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
	response.Success(c, http.StatusOK, "Successfully patched Procedure", result.FullResponse)
}

func (h *ProcedureHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Procedure", result.FullResponse)
}

func (h *ProcedureHandler) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Procedures", result.FullResponse)
}
