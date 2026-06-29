package kyc

import (
	"context"
	"encoding/json"

	"service/internal/interfaces/satusehat"
	"service/pkg/crypto"
	"service/pkg/errors"
)

type Repository interface {
	// Menerima privKeyPEM untuk keperluan decrypt hasil response
	GenerateURL(ctx context.Context, payload GenerateURLRequest, privKeyPEM string) (*satusehat.FHIRResponse, error)
}

type repository struct {
	client satusehat.SatuSehatClient
}

func NewRepository(client satusehat.SatuSehatClient) Repository {
	return &repository{client: client}
}

// hiddenCtx membungkus context untuk mencegah *HTTP client*
// melakukan type-assertion ke *gin.Context dan melakukan auto-print error.
type hiddenCtx struct {
	context.Context
}

func (r *repository) GenerateURL(ctx context.Context, payload GenerateURLRequest, privKeyPEM string) (*satusehat.FHIRResponse, error) {
	// 1. Marshal JSON asli
	jsonPayload, _ := json.Marshal(payload)

	// 2. Enkripsi JSON Payload menggunakan fungsi AES/RSA yang baru dibuat
	encryptedStringPayload, err := crypto.EncryptSatuSehatPayload(jsonPayload)
	if err != nil {
		return nil, errors.InternalError().Message("Gagal mengenkripsi payload").Cause(err).Build()
	}

	// Endpoint ini mengarah ke /generate-url dengan method POST
	// NOTE: Pastikan r.client.DoKYC mendukung pengiriman string dan otomatis
	// memberikan HTTP Header Content-Type: text/plain
	respData, err := r.client.DoKYC(hiddenCtx{ctx}, "POST", "/generate-url", encryptedStringPayload)
	if err != nil {
		return nil, errors.InternalError().
			Message("Failed to generate KYC URL to SatuSehat").
			Cause(err).
			Build()
	}

	// 3. Dekripsi response dari Kemenkes menggunakan Private Key kita
	decryptedData, decErr := crypto.DecryptSatuSehatPayload(string(respData), privKeyPEM)
	if decErr != nil {
		// Terkadang jika validasi NIK gagal, Kemenkes membalas JSON plaintext berisi {"metadata": {"code":"400"...}}
		// Jika gagal decrypt, kita asumsikan Kemenkes mengirim plaintext error, kita parse langsung.
		decryptedData = respData
	}

	var result map[string]interface{}
	if err := json.Unmarshal(decryptedData, &result); err != nil {
		return nil, errors.InternalError().
			Message("Failed to parse KYC response").
			Cause(err).
			Metadata("raw_response", string(decryptedData)).
			Build()
	}

	return &satusehat.FHIRResponse{
		FullResponse: result,
		RawResponse:  respData,
	}, nil
}
