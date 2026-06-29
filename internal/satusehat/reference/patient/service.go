package patient

import (
	"context"
)

// Service mendefinisikan interface untuk manajemen Pasien Satu Sehat
type Service interface {
	GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error)
	GetByID(ctx context.Context, id string) (map[string]interface{}, error)
	Search(ctx context.Context, params PatientSearchParams) (map[string]interface{}, error)
	Create(ctx context.Context, req CreatePatientRequest) (map[string]interface{}, error)
}

type service struct {
	repo Repository
}

// NewService membuat instance baru dari patient service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByNIK(ctx context.Context, nik string) (map[string]interface{}, error) {
	return s.repo.GetByNIK(ctx, nik)
}

func (s *service) GetByID(ctx context.Context, id string) (map[string]interface{}, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Search(ctx context.Context, params PatientSearchParams) (map[string]interface{}, error) {
	return s.repo.Search(ctx, params)
}

func (s *service) Create(ctx context.Context, req CreatePatientRequest) (map[string]interface{}, error) {
	return s.repo.Create(ctx, req)
}
