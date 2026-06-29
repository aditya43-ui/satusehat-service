package bpjs

import (
	"service/internal/infrastructure/config"
	"strings"
)

const (

	// Konstanta Base Service Name (Lingkungan Production)
	ServiceNameAntreanRS   = "antreanrs"
	ServiceNameAntreanFKTP = "antreanfktp"
	ServiceNameApotek      = "apotek-rest"
	ServiceNamePCare       = "pcare-rest"
	ServiceNameICare       = "ihs"
	ServiceNameERekamMedis = "erekammedis"
	ServiceNameAplicare    = "aplicaresws"
	ServiceNameVClaim      = "vclaim-rest"
)

// Factory menyediakan akses tersentralisasi ke berbagai layanan HTTP Client BPJS yang Fleksibel.
type Factory struct {
	baseConfig config.BpjsConfig
}

// NewBPJSFactory membuat instance factory baru dengan konfigurasi utama.
func NewBPJSFactory(cfg config.BpjsConfig) *Factory {
	return &Factory{
		baseConfig: cfg,
	}
}

func (f *Factory) resolveServiceName(baseName string) string {
	name := baseName
	// Auto-detect environment berdasarkan BaseURL (jika URL mengandung kata 'dev')
	if strings.Contains(f.baseConfig.BaseURL, "dev") {
		if baseName == ServiceNameVClaim || baseName == ServiceNameApotek || baseName == ServiceNamePCare {
			name = baseName + "-dev"
		} else if baseName == ServiceNameAplicare {
			name = baseName // Pengecualian: Aplicares sama sekali tidak menggunakan akhiran -dev atau _dev
		} else {
			name = baseName + "_dev" // AntreanRS, dll menggunakan underscore
		}
	}
	return name
}

func (f *Factory) createClient(serviceName string, customConsID string, customSecretKey string, customUserKey string) BpjsClient {
	svcConfig := f.baseConfig

	// Timpa kredensial Default dengan kredensial spesifik layanan (Khusus Apotek)
	if customConsID != "" {
		svcConfig.ConsID = customConsID
	}
	if customSecretKey != "" {
		svcConfig.SecretKey = customSecretKey
	}
	// Timpa Active UserKey
	svcConfig.UserKey = customUserKey
	// [HOTFIX] Penyesuaian Base URL khusus untuk Aplicares
	// Karena Aplicares BPJS (Ketersediaan Tempat Tidur) BELUM dimigrasikan ke API Gateway Kong (apijkn).
	// Aplicares masih menggunakan server lamanya yaitu dvlp (Dev) dan new-api (Prod).
	if serviceName == ServiceNameAplicare {
		if strings.Contains(svcConfig.BaseURL, "apijkn-dev") {
			svcConfig.BaseURL = "https://dvlp.bpjs-kesehatan.go.id:8888"
		} else if strings.Contains(svcConfig.BaseURL, "apijkn.bpjs-kesehatan.go.id") {
			svcConfig.BaseURL = "https://new-api.bpjs-kesehatan.go.id"
		}
	}
	svcConfig.ServiceName = f.resolveServiceName(serviceName)
	return NewBpjsClient(svcConfig)
}

// --- Kumpulan Instance Client untuk Masing-masing Layanan BPJS ---

func (f *Factory) VClaim() BpjsClient {
	// Gunakan "" agar otomatis memakai kredensial Default (VClaim)
	return f.createClient(ServiceNameVClaim, "", "", f.baseConfig.VclaimUserKey)
}

func (f *Factory) AntreanRS() BpjsClient {
	return f.createClient(ServiceNameAntreanRS, "", "", f.baseConfig.AntrolUserKey)
}

func (f *Factory) AntreanFKTP() BpjsClient {
	return f.createClient(ServiceNameAntreanFKTP, "", "", f.baseConfig.AntrolUserKey)
}

func (f *Factory) Aplicare() BpjsClient {
	return f.createClient(ServiceNameAplicare, "", "", f.baseConfig.AplicareUserKey)
}

func (f *Factory) Apotek() BpjsClient {
	// Apotek menggunakan ConsID dan SecretKey nya sendiri!
	return f.createClient(ServiceNameApotek, f.baseConfig.ApotekConsID, f.baseConfig.ApotekSecretKey, f.baseConfig.ApotekUserKey)
}

func (f *Factory) PCare() BpjsClient {
	return f.createClient(ServiceNamePCare, "", "", "")
}

func (f *Factory) ICare() BpjsClient {
	return f.createClient(ServiceNameICare, "", "", f.baseConfig.IhsUserKey)
}

func (f *Factory) ERekamMedis() BpjsClient {
	return f.createClient(ServiceNameERekamMedis, "", "", f.baseConfig.IhsUserKey)
}
