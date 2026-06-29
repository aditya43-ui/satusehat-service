package errors

import (
	"strings"
)

// ParseSatuSehatError mengekstrak dan menerjemahkan balasan OperationOutcome
// dari Kemenkes (SatuSehat) menjadi standard AppError aplikasi.
func ParseSatuSehatError(result map[string]interface{}) Error {
	errMsg := "Gagal memproses data ke SatuSehat"
	var errMessages []string

	// Ekstrak pesan dari array "issue"
	if issues, ok := result["issue"].([]interface{}); ok && len(issues) > 0 {
		for _, issueItem := range issues {
			if issue, ok := issueItem.(map[string]interface{}); ok {
				if diagnostics, ok := issue["diagnostics"].(string); ok {
					diagLower := strings.ToLower(diagnostics)

					// Translasi pesan error spesifik SatuSehat
					if strings.Contains(diagLower, "reference target(s) not found") {
						errMessages = append(errMessages, "Data referensi (seperti ID) tidak ditemukan di SatuSehat")
					} else if strings.Contains(diagLower, "nik") {
						errMessages = append(errMessages, "NIK tidak valid atau tidak terdaftar di SatuSehat")
					} else {
						errMessages = append(errMessages, diagnostics) // Tampilkan pesan aslinya jika belum ada di kamus translasi
					}
				}
			}
		}
	}

	// Gabungkan semua error (jika > 1) dengan pemisah koma atau titik koma
	if len(errMessages) > 0 {
		errMsg = "Validasi SatuSehat: " + strings.Join(errMessages, "; ")
	}

	// Gunakan NewValidationError agar status HTTP menjadi 400 (Bad Request)
	return NewValidationError().
		Message(errMsg).
		Metadata("operation_outcome", result).
		Build()
}
