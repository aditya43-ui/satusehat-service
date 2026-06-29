package kfa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
	"service/pkg/logger"
)

type Repository interface {
	GetByCode(ctx context.Context, code string) (map[string]interface{}, error)
	GetProducts(ctx context.Context, params KFASearchParams) (map[string]interface{}, error)
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
		return nil, fmt.Errorf("failed to parse KFA response: %w", err)
	}
	return res, nil
}

func (r *repository) GetByCode(ctx context.Context, code string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/products?identifier=kfa&code=%s", url.QueryEscape(code))
	respBytes, err := r.client.DoKFA(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari produk KFA berdasarkan kode", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mencari produk KFA").Cause(err).Build()
	}
	return parseResponse(respBytes)
}

func (r *repository) GetProducts(ctx context.Context, params KFASearchParams) (map[string]interface{}, error) {
	q := url.Values{}
	q.Add("page", strconv.Itoa(params.Page))
	q.Add("size", strconv.Itoa(params.Size))
	if params.ProductType != "" {
		q.Add("product_type", params.ProductType)
	}
	if params.Keyword != "" {
		q.Add("keyword", params.Keyword)
	}
	if params.From != "" {
		q.Add("from_", params.From)
	}

	endpoint := "/products/all?" + q.Encode()
	respBytes, err := r.client.DoKFA(ctx, "GET", endpoint, nil)
	if err != nil {
		logger.Default().Error("Gagal mencari daftar produk KFA", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mengambil daftar produk KFA").Cause(err).Build()
	}
	return parseResponse(respBytes)
}
