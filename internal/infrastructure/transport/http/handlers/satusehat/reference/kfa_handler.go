package reference

import (
	"net/http"

	"service/internal/satusehat/reference/kfa"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type KFAHandler struct {
	service kfa.Service
}

func NewKFAHandler(service kfa.Service) *KFAHandler {
	return &KFAHandler{
		service: service,
	}
}

// GetByCode godoc
//
//	@Summary		Cari Produk KFA berdasarkan Kode
//	@Description	Mencari detail produk farmasi/alkes dari API Kamus Farmasi dan Alat Kesehatan (KFA) Satu Sehat
//	@Tags			Satu Sehat - KFA
//	@Produce		json
//	@Param			code	path		string	true	"Kode KFA (contoh: 93000469)"
//	@Success		200		{object}	response.Response
//	@Router			/satusehat/reference/kfa/products/{code} [get]
//	@Security		BearerAuth
func (h *KFAHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	data, err := h.service.GetByCode(c.Request.Context(), code)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan data produk KFA", data)
}

// GetProducts godoc
//
//	@Summary		Daftar Produk KFA
//	@Description	Mengambil daftar produk KFA dengan kapabilitas paginasi
//	@Tags			Satu Sehat - KFA
//	@Produce		json
//	@Param			page			query		int		false	"Nomor Halaman (default: 1)"
//	@Param			size			query		int		false	"Jumlah Data (default: 10)"
//	@Param			product_type	query		string	false	"Tipe Produk (farmasi/alkes)"
//	@Param			keyword			query		string	false	"Kata Kunci Pencarian"
//	@Param			from_			query		string	false	"Parameter waktu (from_)"
//	@Success		200				{object}	response.Response
//	@Router			/satusehat/reference/kfa/products [get]
//	@Security		BearerAuth
func (h *KFAHandler) GetProducts(c *gin.Context) {
	var params kfa.KFASearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Error(c, http.StatusBadRequest, "Format parameter tidak valid", err.Error())
		return
	}

	data, err := h.service.GetProducts(c.Request.Context(), params)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mengambil daftar produk KFA", data)
}

func (h *KFAHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference/kfa")
	{
		group.GET("/products/:code", h.GetByCode)
		group.GET("/products", h.GetProducts)
	}
}
