package reference

import (
	"net/http"

	"service/internal/satusehat/reference/practitioner"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type PractitionerHandler struct {
	service practitioner.Service
}

func NewPractitionerHandler(service practitioner.Service) *PractitionerHandler {
	return &PractitionerHandler{
		service: service,
	}
}

// GetByNIK godoc
//
//	@Summary		Cari Tenaga Medis (Satu Sehat) berdasarkan NIK
//	@Description	Mencari data Practitioner FHIR Satu Sehat berdasarkan NIK
//	@Tags			Satu Sehat - Practitioner
//	@Produce		json
//	@Param			nik	path		string	true	"Nomor Induk Kependudukan"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/practitioner/nik/{nik} [get]
//	@Security		BearerAuth
func (h *PractitionerHandler) GetByNIK(c *gin.Context) {
	nik := c.Param("nik")
	data, err := h.service.GetByNIK(c.Request.Context(), nik)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data tenaga medis", data)
}

// GetByID godoc
//
//	@Summary		Cari Tenaga Medis (Satu Sehat) berdasarkan ID
//	@Description	Mencari data Practitioner FHIR Satu Sehat berdasarkan IHS Number / ID
//	@Tags			Satu Sehat - Practitioner
//	@Produce		json
//	@Param			id	path		string	true	"IHS Number / Practitioner ID"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/practitioner/{id} [get]
//	@Security		BearerAuth
func (h *PractitionerHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data tenaga medis", data)
}

// Search godoc
//
//	@Summary		Pencarian Tenaga Medis (Satu Sehat) Multi-Parameter
//	@Description	Mencari data Practitioner FHIR Satu Sehat berdasarkan parameter
//	@Tags			Satu Sehat - Practitioner
//	@Produce		json
//	@Param			nik			query		string	false	"Nomor Induk Kependudukan"
//	@Param			name		query		string	false	"Nama Tenaga Medis"
//	@Param			gender		query		string	false	"Jenis Kelamin (male/female)"
//	@Param			birthdate	query		string	false	"Tanggal Lahir (YYYY-MM-DD)"
//	@Success		200			{object}	response.Response
//	@Router			/satusehat/reference/practitioner [get]
//	@Security		BearerAuth
func (h *PractitionerHandler) Search(c *gin.Context) {
	var params practitioner.PractitionerSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Error(c, http.StatusBadRequest, "Format pencarian tidak valid", err.Error())
		return
	}

	data, err := h.service.Search(c.Request.Context(), params)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data tenaga medis", data)
}

// RegisterRoutes mendaftarkan endpoint handler ini ke router Gin
func (h *PractitionerHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference")
	{
		group.GET("/practitioner/nik/:nik", h.GetByNIK)
		group.GET("/practitioner/:id", h.GetByID)
		group.GET("/practitioner", h.Search)
	}
}
