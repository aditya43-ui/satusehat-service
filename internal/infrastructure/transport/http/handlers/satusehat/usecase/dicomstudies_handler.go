package usecase

import (
	"io"
	"net/http"

	"service/internal/satusehat/usecase/studies"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type DicomStudiesHandler struct {
	service studies.Service
}

func NewDicomStudiesHandler(service studies.Service) *DicomStudiesHandler {
	return &DicomStudiesHandler{
		service: service,
	}
}

func (h *DicomStudiesHandler) RegisterRoutes(router *gin.RouterGroup) {
	// URL akan menjadi: POST /api/v1/satusehat/dicom/studies
	group := router.Group("/satusehat/dicom/studies")
	{
		group.POST("", h.Upload)
	}
}

func (h *DicomStudiesHandler) Upload(c *gin.Context) {
	// Ambil file dari request FormData dengan key "file"
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File DICOM tidak ditemukan pada form-data key 'file'", err.Error())
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuka file DICOM sementara", err.Error())
		return
	}
	defer openedFile.Close()

	dicomBytes, err := io.ReadAll(openedFile)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membaca byte dari file DICOM", err.Error())
		return
	}

	result, err := h.service.UploadDICOM(c.Request.Context(), dicomBytes)
	if err != nil {
		// Fungsi ini berasal dari encounter_handler.go di package yang sama
		handleSatuSehatError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Berhasil mengunggah file DICOM ke SatuSehat STOW-RS", result.FullResponse)
}
