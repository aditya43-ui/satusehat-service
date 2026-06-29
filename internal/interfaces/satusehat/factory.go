package satusehat

import (
	"service/internal/infrastructure/config"
)

// Factory menyediakan akses tersentralisasi ke HTTP Client SatuSehat.
type Factory struct {
	baseConfig config.SatuSehatConfig
}

// NewSatuSehatFactory membuat instance factory baru untuk SatuSehat.
func NewSatuSehatFactory(cfg config.SatuSehatConfig) *Factory {
	return &Factory{
		baseConfig: cfg,
	}
}

// Client mengembalikan instance utama dari SatuSehatClient.
// Jika kelak dibutuhkan client dengan konfigurasi khusus atau environment berbeda,
// pembuatannya dapat dipusatkan di Factory ini.
func (f *Factory) Client() SatuSehatClient {
	return NewSatuSehatClient(f.baseConfig)
}
