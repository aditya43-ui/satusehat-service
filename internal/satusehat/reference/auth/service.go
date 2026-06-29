package auth

import (
	"context"
	"encoding/json"
	"time"

	"service/internal/infrastructure/cache"
	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
	"service/pkg/logger"
)

// Service mendefinisikan interface untuk manajemen Auth Satu Sehat di level aplikasi.
type Service interface {
	GetToken(ctx context.Context) (map[string]interface{}, error)
	RefreshToken(ctx context.Context) (map[string]interface{}, error)
}

type service struct {
	client satusehat.SatuSehatClient
	cache  *cache.Manager
}

// NewService membuat instance baru dari auth service.
func NewService(client satusehat.SatuSehatClient, cacheManager *cache.Manager) Service {
	return &service{
		client: client,
		cache:  cacheManager,
	}
}

func (s *service) GetToken(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "satusehat:auth_data"

	// 1. Coba ambil dari Redis cache (sangat efisien untuk multi-instance/load balancer)
	if s.cache != nil {
		var cachedToken string
		if err := s.cache.Get(ctx, cacheKey, &cachedToken); err == nil && cachedToken != "" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(cachedToken), &data); err == nil {
				logger.Default().Debug("🔑 Mendapatkan token Satu Sehat dari Redis Cache")
				return data, nil
			}
		}
	}

	// 2. Jika tidak ada di Redis (atau expired), ambil dari Klien Kemenkes
	data, err := s.client.GetAccessToken(ctx)
	if err != nil {
		logger.Default().Error("Gagal mendapatkan token dari SatuSehat Client", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal mendapatkan token Satu Sehat dari Kemenkes").Cause(err).Build()
	}

	// 3. Simpan ke Redis cache (Diset 55 menit karena token Kemenkes biasanya expired dalam 60 menit)
	if s.cache != nil {
		if bytes, err := json.Marshal(data); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 55*time.Minute)
		}
	}

	return data, nil
}

func (s *service) RefreshToken(ctx context.Context) (map[string]interface{}, error) {
	// Refresh token secara paksa melewati cache in-memory HTTP Client
	data, err := s.client.RefreshToken(ctx)
	if err != nil {
		logger.Default().Error("Gagal me-refresh token SatuSehat Client", logger.ErrorField(err))
		return nil, errors.InternalError().Message("Gagal melakukan refresh token Satu Sehat").Cause(err).Build()
	}

	// Update sinkronisasi token terbaru ke Redis cache
	if s.cache != nil {
		cacheKey := "satusehat:auth_data"
		if bytes, err := json.Marshal(data); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 55*time.Minute)
		}
	}

	logger.Default().Info("🔄 Token Satu Sehat berhasil di-refresh secara manual")
	return data, nil
}
