package location

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
	Search(ctx context.Context, params LocationSearchParams) (map[string]interface{}, error)
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
	endpoint := fmt.Sprintf("/Location/%s", id)
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Location berdasarkan ID", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari Location").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Search(ctx context.Context, params LocationSearchParams) (map[string]interface{}, error) {
	q := url.Values{}
	if params.Name != "" {
		q.Add("name", params.Name)
	}
	if params.Organization != "" {
		q.Add("organization", params.Organization)
	}
	if params.Identifier != "" {
		q.Add("identifier", params.Identifier)
	}

	endpoint := "/Location?" + q.Encode()
	respBytes, err := r.client.DoRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari Location", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal melakukan pencarian Location").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) Create(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "POST", "/Location", payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}

func (r *repository) Update(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "PUT", fmt.Sprintf("/Location/%s", id), payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}
func (r *repository) Patch(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	respBytes, err := r.client.DoRequest(ctx, "PATCH", fmt.Sprintf("/Location/%s", id), payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(respBytes)
}
