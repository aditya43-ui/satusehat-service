package reference

import (
	"net/http"

	"service/internal/satusehat/reference/location"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	service location.Service
}

func NewLocationHandler(service location.Service) *LocationHandler {
	return &LocationHandler{
		service: service,
	}
}

// GetByID godoc
//
//	@Summary		Cari Lokasi (Satu Sehat) berdasarkan ID
//	@Description	Mencari data Location FHIR Satu Sehat berdasarkan ID
//	@Tags			Satu Sehat - Location
//	@Produce		json
//	@Param			id	path		string	true	"Location ID"
//	@Success		200	{object}	response.Response
//	@Router			/satusehat/reference/location/{id} [get]
//	@Security		BearerAuth
func (h *LocationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data lokasi", data)
}

// Search godoc
//
//	@Summary		Pencarian Lokasi (Satu Sehat)
//	@Description	Mencari data Location berdasarkan parameter
//	@Tags			Satu Sehat - Location
//	@Produce		json
//	@Param			name			query		string	false	"Nama Lokasi"
//	@Param			organization	query		string	false	"ID Organisasi Pemilik"
//	@Param			identifier		query		string	false	"Identifier Lokasi"
//	@Success		200				{object}	response.Response
//	@Router			/satusehat/reference/location [get]
//	@Security		BearerAuth
func (h *LocationHandler) Search(c *gin.Context) {
	var params location.LocationSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Error(c, http.StatusBadRequest, "Format pencarian tidak valid", err.Error())
		return
	}

	data, err := h.service.Search(c.Request.Context(), params)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data lokasi", data)
}

func (h *LocationHandler) Create(c *gin.Context) {
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
	response.Success(c, http.StatusCreated, "Berhasil membuat data lokasi", data)
}

func (h *LocationHandler) Update(c *gin.Context) {
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
	response.Success(c, http.StatusOK, "Berhasil mengubah data lokasi", data)
}

func (h *LocationHandler) Patch(c *gin.Context) {
	id := c.Param("id")
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
	response.Success(c, http.StatusOK, "Berhasil melakukan patch data lokasi", data)
}

// RegisterRoutes mendaftarkan endpoint handler ini ke router Gin
func (h *LocationHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference")
	{
		group.GET("/location/:id", h.GetByID)
		group.GET("/location", h.Search)
		group.POST("/location", h.Create)
		group.PUT("/location/:id", h.Update)
		group.PATCH("/location/:id", h.Patch)
	}
}
