package usecase

import (
	"net/http"
	"service/internal/satusehat/usecase/diagnosticreport"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type DiagnosticReportHandler struct{ service diagnosticreport.Service }

func NewDiagnosticReportHandler(s diagnosticreport.Service) *DiagnosticReportHandler {
	return &DiagnosticReportHandler{service: s}
}

func (h *DiagnosticReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/satusehat/diagnosticreport")
	g.POST("", h.Create)
	g.GET("", h.Search)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id", h.Patch)
}

func (h *DiagnosticReportHandler) Create(c *gin.Context) {
	var req diagnosticreport.DiagnosticReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format payload JSON tidak valid", validator.TranslateError(err))
		return
	}
	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Success", res.FullResponse)
}
func (h *DiagnosticReportHandler) Update(c *gin.Context) {
	var req diagnosticreport.DiagnosticReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format payload JSON tidak valid", validator.TranslateError(err))
		return
	}
	res, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *DiagnosticReportHandler) Patch(c *gin.Context) {
	var req diagnosticreport.DiagnosticReportPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format patch payload tidak valid", validator.TranslateError(err))
		return
	}
	res, err := h.service.Patch(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *DiagnosticReportHandler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
func (h *DiagnosticReportHandler) Search(c *gin.Context) {
	res, err := h.service.Search(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Success", res.FullResponse)
}
