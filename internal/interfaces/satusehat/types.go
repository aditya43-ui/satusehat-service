package satusehat

import (
	"context"
	"encoding/json"
	"fmt"
)

// SatuSehatClient adalah core HTTP client untuk API SatuSehat (Kemenkes).
type SatuSehatClient interface {
	// GetAccessToken mengambil access token aktif (dari memori atau request baru jika kadaluwarsa).
	GetAccessToken(ctx context.Context) (map[string]interface{}, error)
	// RefreshToken memaksa pengambilan access token baru dari server Kemenkes.
	RefreshToken(ctx context.Context) (map[string]interface{}, error)
	// DoRequest mengirimkan HTTP request ke endpoint utama FHIR (BaseURL).
	DoRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error)
	// DoKFA mengirimkan HTTP request ke endpoint KFA (KFAURL).
	DoKFA(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error)
	// DoConsent mengirimkan HTTP request ke endpoint Consent (ConsentURL).
	DoConsent(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error)
	// DoKYC mengirimkan HTTP request ke endpoint KYC (KYCURL).
	DoKYC(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error)
	// UploadDICOM mengirimkan file biner .dcm menggunakan protokol STOW-RS (multipart/related).
	UploadDICOM(ctx context.Context, dicomBytes []byte) ([]byte, error)
}

// FHIRPayload adalah struktur data fleksibel berbentuk map untuk membangun JSON payload FHIR Satu Sehat.
type FHIRPayload map[string]interface{}

// NewFHIRPayload membuat instance FHIRPayload baru dengan `resourceType` yang selalu diwajibkan oleh standar FHIR.
func NewFHIRPayload(resourceType string) FHIRPayload {
	return FHIRPayload{
		"resourceType": resourceType,
	}
}

// Set menambahkan atau menimpa nilai (value) pada JSON berdasarkan key.
// Menggunakan fluent interface (mengembalikan FHIRPayload) agar bisa di-chain (disambung).
func (payload FHIRPayload) Set(key string, value interface{}) FHIRPayload {
	payload[key] = value
	return payload
}

// Append menambahkan nilai (value) ke dalam sebuah array JSON berdasarkan key.
// Sangat berguna untuk FHIR yang sangat sering menggunakan array (seperti `identifier`, `name`, `telecom`).
// Jika key belum ada, ia akan secara otomatis membuat array baru.
func (payload FHIRPayload) Append(arrayKey string, value interface{}) FHIRPayload {
	if existing, ok := payload[arrayKey]; ok {
		// Jika array sudah ada, tambahkan value baru ke dalamnya
		if slice, isSlice := existing.([]interface{}); isSlice {
			payload[arrayKey] = append(slice, value)
		} else {
			payload[arrayKey] = []interface{}{existing, value} // Fallback safety
		}
	} else {
		// Jika array belum ada, buat array baru dengan 1 elemen
		payload[arrayKey] = []interface{}{value}
	}
	return payload
}

// ToJSON menghasilkan representasi byte JSON dari payload yang telah dibangun.
func (payload FHIRPayload) ToJSON() ([]byte, error) {
	return json.Marshal(payload)
}

// ErrorOperationOutcome merepresentasikan error terstruktur dari API SatuSehat (FHIR OperationOutcome)
type ErrorOperationOutcome struct {
	StatusCode int
	Outcome    map[string]interface{}
}

func (err *ErrorOperationOutcome) Error() string {
	return fmt.Sprintf("SatuSehat API Error (HTTP %d)", err.StatusCode)
}

// FHIRResponse represents a structured response from the SatuSehat API.
type FHIRResponse struct {
	// The ID of the created or retrieved FHIR resource.
	ID string `json:"id"`
	// The full JSON response from the API as a map.
	FullResponse map[string]interface{} `json:"full_response"`
	// The raw JSON response from the API as a byte slice.
	RawResponse []byte `json:"-"` // Ignored by JSON marshalling for client response
}
