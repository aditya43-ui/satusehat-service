package organization

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
	GetByID(ctx context.Context, id string) (map[string]interface{}, error)
	Search(ctx context.Context, params OrganizationSearchParams) (map[string]interface{}, error)
	Create(ctx context.Context, payload interface{}) (map[string]interface{}, error)
	Update(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error)
	Patch(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error)
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

func (r *repository) GetByID(ctx context.Context, id string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Organization/%s", id)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Organization berdasarkan ID", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Organization").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Search(ctx context.Context, params OrganizationSearchParams) (map[string]interface{}, error) {
	q := url.Values{}
	if params.Name != "" {
		q.Add("name", params.Name)
	}
	if params.PartOf != "" {
		q.Add("partof", params.PartOf)
	}
	if params.Identifier != "" {
		q.Add("identifier", params.Identifier)
	}

	endpoint := "/Organization?" + q.Encode()
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Organization", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal melakukan pencarian Organization").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Create(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "POST", "/Organization", payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}

func (r *repository) Update(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "PUT", fmt.Sprintf("/Organization/%s", id), payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}

func (r *repository) Patch(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "PATCH", fmt.Sprintf("/Organization/%s", id), payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}
