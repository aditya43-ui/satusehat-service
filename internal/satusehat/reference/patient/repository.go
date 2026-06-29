package patient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
	"service/pkg/logger"
)

// Repository mendefinisikan antarmuka komunikasi ke eksternal (Kemenkes API).
type Repository interface {
	GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error)
	GetByID(ctx context.Context, id string) (map[string]interface{}, error)
	Search(ctx context.Context, params PatientSearchParams) (map[string]interface{}, error)
	Create(ctx context.Context, req CreatePatientRequest) (map[string]interface{}, error)
}

type repository struct {
	client satusehat.SatuSehatClient
}

// NewRepository membuat instance baru dari patient repository.
func NewRepository(client satusehat.SatuSehatClient) Repository {
	return &repository{client: client}
}

// helper untuk mem-parsing response body []byte menjadi Map JSON
func parseResponse(body []byte) (map[string]interface{}, error) {
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to parse SatuSehat response: %w", err)
	}
	return res, nil
}

func (r *repository) GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Patient?identifier=%s|%s", IdentifierSystemNIK, nik)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Pasien berdasarkan NIK", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Pasien ke Satu Sehat").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) GetByID(ctx context.Context, id string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Patient/%s", id)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Pasien berdasarkan ID", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Pasien ke Satu Sehat").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Search(ctx context.Context, params PatientSearchParams) (map[string]interface{}, error) {
	q := url.Values{}

	if params.Name != "" {
		q.Add("name", params.Name)
	}
	if params.BirthDate != "" {
		q.Add("birthdate", params.BirthDate)
	}
	if params.Gender != "" {
		q.Add("gender", params.Gender)
	}
	if params.NIK != "" {
		q.Add("identifier", IdentifierSystemNIK+"|"+params.NIK)
	} else if params.NIKIbu != "" {
		q.Add("identifier", IdentifierSystemNIKIbu+"|"+params.NIKIbu)
	}

	endpoint := "/Patient?" + q.Encode()
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Pasien dengan parameter dinamis", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal melakukan pencarian Pasien ke Satu Sehat").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Create(ctx context.Context, req CreatePatientRequest) (map[string]interface{}, error) {
	payload := satusehat.NewFHIRPayload("Patient").
		Set("active", true).
		Set("gender", req.Gender).
		Set("birthDate", req.BirthDate).
		Append("identifier", map[string]interface{}{
			"use":    "official",
			"system": IdentifierSystemNIK,
			"value":  req.NIK,
		}).
		Append("name", map[string]interface{}{
			"use":  "official",
			"text": req.Name,
		})

	if req.Phone != "" {
		payload.Append("telecom", map[string]interface{}{
			"system": "phone",
			"value":  req.Phone,
			"use":    "mobile",
		})
	}
	if req.Address != "" {
		payload.Append("address", map[string]interface{}{
			"use":  "home",
			"text": req.Address,
		})
	}

	respBytes, err := r.client.DoRequest(ctx, "POST", "/Patient", payload)
	if err != nil {
		logger.Default().Error("Gagal mendaftarkan Pasien", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mendaftarkan Pasien ke Satu Sehat").Cause(err).Build()
	}

	return parseResponse(respBytes)
}
