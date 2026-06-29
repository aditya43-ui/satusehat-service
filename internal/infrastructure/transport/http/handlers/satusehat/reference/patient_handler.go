package reference

import (
	"net/http"
	"service/internal/satusehat/reference/patient"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

// PatientHandler menangani endpoint HTTP untuk resource Patient Satu Sehat.
type PatientHandler struct {
	service patient.Service
}

// RegisterRoutes mendaftarkan endpoint handler ini ke router Gin
func (h *PatientHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference")
	{
		group.GET("/patient/nik/:nik", h.GetByNIK)
		group.GET("/patient/:id", h.GetByID)
		group.GET("/patient", h.Search)
		group.POST("/patient", h.Create)
	}
}

// NewPatientHandler membuat instance baru dari PatientHandler.
func NewPatientHandler(service patient.Service) *PatientHandler {
	return &PatientHandler{
		service: service,
	}
}

// GetByNIK godoc
//
//	@Summary		Cari Pasien (Satu Sehat) berdasarkan NIK
//	@Description	Mencari data pasien FHIR Satu Sehat berdasarkan Nomor Induk Kependudukan (NIK)
//	@Tags			Satu Sehat - Patient
//	@Produce		json
//	@Param			nik	path		string	true	"Nomor Induk Kependudukan"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/patient/nik/{nik} [get]
//	@Security		BearerAuth
func (h *PatientHandler) GetByNIK(c *gin.Context) {
	nik := c.Param("nik")
	data, err := h.service.GetByNIK(c.Request.Context(), nik)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data pasien", data)
}

// GetByID godoc
//
//	@Summary		Cari Pasien (Satu Sehat) berdasarkan ID
//	@Description	Mencari data pasien FHIR Satu Sehat berdasarkan IHS Number / ID
//	@Tags			Satu Sehat - Patient
//	@Produce		json
//	@Param			id	path		string	true	"IHS Number / Patient ID"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/patient/{id} [get]
//	@Security		BearerAuth
func (h *PatientHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data pasien", data)
}

// Search godoc
//
//	@Summary		Pencarian Pasien (Satu Sehat) Multi-Parameter
//	@Description	Mencari data pasien FHIR Satu Sehat berdasarkan kombinasi parameter (Nama, NIK, NIK Ibu, Tanggal Lahir, Gender)
//	@Tags			Satu Sehat - Patient
//	@Produce		json
//	@Param			nik			query		string	false	"Nomor Induk Kependudukan"
//	@Param			nik_ibu		query		string	false	"Nomor Induk Kependudukan Ibu (Untuk bayi)"
//	@Param			name		query		string	false	"Nama Pasien"
//	@Param			birthdate	query		string	false	"Tanggal Lahir (YYYY-MM-DD)"
//	@Param			gender		query		string	false	"Jenis Kelamin (male/female)"
//	@Success		200			{object}	response.Response
//	@Router			/satusehat/reference/patient [get]
//	@Security		BearerAuth
func (h *PatientHandler) Search(c *gin.Context) {
	var params patient.PatientSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Error(c, http.StatusBadRequest, "Format pencarian tidak valid", err.Error())
		return
	}

	data, err := h.service.Search(c.Request.Context(), params)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data pasien", data)
}

// Create godoc
//
//	@Summary		Daftar Pasien Baru (Satu Sehat)
//	@Description	Mendaftarkan data pasien baru ke API Satu Sehat dan mengembalikan IHS Number
//	@Tags			Satu Sehat - Patient
//	@Accept			json
//	@Produce		json
//	@Param			request	body		patient.CreatePatientRequest	true	"Data Pasien Baru"
//	@Success		201		{object}	response.Response
//	@Router			/satusehat/reference/patient [post]
//	@Security		BearerAuth
func (h *PatientHandler) Create(c *gin.Context) {
	var req patient.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Format request tidak valid", err.Error())
		return
	}

	data, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Berhasil mendaftarkan pasien", data)
}
