package reference

import (
	"net/http"

	"service/internal/satusehat/reference/organization"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrganizationHandler struct {
	service organization.Service
}

func NewOrganizationHandler(service organization.Service) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
	}
}

// GetByID godoc
//
//	@Summary		Cari Organisasi (Satu Sehat) berdasarkan ID
//	@Description	Mencari data Organization FHIR Satu Sehat berdasarkan ID
//	@Tags			Satu Sehat - Organization
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/organization/{id} [get]
//	@Security		BearerAuth
func (h *OrganizationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data organisasi", data)
}

// Search godoc
//
//	@Summary		Pencarian Organisasi (Satu Sehat)
//	@Description	Mencari data Organization berdasarkan Name atau PartOf
//	@Tags			Satu Sehat - Organization
//	@Produce		json
//	@Param			name		query		string	false	"Nama Organisasi"
//	@Param			partof		query		string	false	"ID Organisasi Induk"
//	@Param			identifier	query		string	false	"Identifier Organisasi"
//	@Success		200			{object}	response.Response
//	@Router			/satusehat/reference/organization [get]
//	@Security		BearerAuth
func (h *OrganizationHandler) Search(c *gin.Context) {
	var params organization.OrganizationSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Error(c, http.StatusBadRequest, "Format pencarian tidak valid", err.Error())
		return
	}

	// Validasi mandiri sebelum menembak ke API Kemenkes
	if params.Name == "" && params.PartOf == "" && params.Identifier == "" {
		response.Error(c, http.StatusBadRequest, "Parameter pencarian tidak lengkap", "Harap masukkan minimal salah satu parameter query: name, partof, atau identifier")
		return
	}

	data, err := h.service.Search(c.Request.Context(), params)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data organisasi", data)
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "Format request tidak valid", err.Error())
		return
	}

	data, err := h.service.Create(c.Request.Context(), payload)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Berhasil membuat data organisasi", data)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "Format request tidak valid", err.Error())
		return
	}

	data, err := h.service.Update(c.Request.Context(), id, payload)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mengubah data organisasi", data)
}

func (h *OrganizationHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	// JSON Patch biasanya array of objects, jadi gunakan interface{} general
	var payload interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "Format request tidak valid", err.Error())
		return
	}

	data, err := h.service.Patch(c.Request.Context(), id, payload)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil melakukan patch data organisasi", data)
}

// RegisterRoutes mendaftarkan endpoint handler ini ke router Gin
func (h *OrganizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference")
	{
		group.GET("/organization/:id", h.GetByID)
		group.GET("/organization", h.Search)
		group.POST("/organization", h.Create)
		group.PUT("/organization/:id", h.Update)
		group.PATCH("/organization/:id", h.Patch)
	}
}
