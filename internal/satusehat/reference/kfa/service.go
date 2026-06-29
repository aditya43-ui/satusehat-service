package kfa

import "context"

type Service interface {
	GetByCode(ctx context.Context, code string) (map[string]interface{}, error)
	GetProducts(ctx context.Context, params KFASearchParams) (map[string]interface{}, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByCode(ctx context.Context, code string) (map[string]interface{}, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *service) GetProducts(ctx context.Context, params KFASearchParams) (map[string]interface{}, error) {
	return s.repo.GetProducts(ctx, params)
}
