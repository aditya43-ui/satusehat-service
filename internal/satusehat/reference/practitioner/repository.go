package practitioner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
	"service/pkg/logger"
)

type Repository interface {
	GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error)
	GetByID(ctx context.Context, id string) (map[string]interface{}, error)
	Search(ctx context.Context, params PractitionerSearchParams) (map[string]interface{}, error)
}

type repository struct {
	client satusehat.SatuSehatClient
}

func NewRepository(client satusehat.SatuSehatClient) Repository {
	return &repository{client: client}
}

func parseResponse(body []byte) (map[string]interface{}, error) {
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to parse SatuSehat response: %w", err)
	}
	return res, nil
}

func (r *repository) GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Practitioner?identifier=%s|%s", IdentifierSystemNIK, nik)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Practitioner berdasarkan NIK", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Practitioner ke Satu Sehat").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) GetByID(ctx context.Context, id string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Practitioner/%s", id)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Practitioner berdasarkan ID", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Practitioner ke Satu Sehat").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Search(ctx context.Context, params PractitionerSearchParams) (map[string]interface{}, error) {
	q := url.Values{}
	if params.Name != "" {
		q.Add("name", params.Name)
	}
	if params.Gender != "" {
		q.Add("gender", params.Gender)
	}
	if params.BirthDate != "" {
		q.Add("birthdate", params.BirthDate)
	}
	if params.NIK != "" {
		q.Add("identifier", IdentifierSystemNIK+"|"+params.NIK)
	}

	endpoint := "/Practitioner?" + q.Encode()
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Practitioner", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal melakukan pencarian Practitioner").Cause(err).Build()
	}
	return parseResponse(respBytes)
}
