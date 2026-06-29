package kfa

// KFASearchParams menampung parameter pencarian produk KFA
type KFASearchParams struct {
	Page        int    `form:"page,default=1"`
	Size        int    `form:"size,default=10"`
	ProductType string `form:"product_type"` // Contoh: farmasi, alkes
	Keyword     string `form:"keyword"`      // Kata kunci pencarian produk (jika didukung KFA)
	From        string `form:"from_"`        // Query param from_ (digunakan untuk memfilter tanggal/waktu spesifik)
}
