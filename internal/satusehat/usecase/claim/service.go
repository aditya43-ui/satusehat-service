package claim

import (
	"context"
	"net/url"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, req ClaimRequest) (*satusehat.FHIRResponse, error)
	Update(ctx context.Context, id string, req ClaimRequest) (*satusehat.FHIRResponse, error)
	GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error)
	Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error)
}

type service struct {
	repo  Repository
	orgID string
}

func NewService(repo Repository, orgID string) Service {
	return &service{repo: repo, orgID: orgID}
}

func (s *service) Create(ctx context.Context, req ClaimRequest) (*satusehat.FHIRResponse, error) {
	s.applyDefaults(&req)
	return s.repo.Create(ctx, MapRequestToFHIR(req))
}

func (s *service) Update(ctx context.Context, id string, req ClaimRequest) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("Claim ID is required").Build()
	}
	s.applyDefaults(&req)
	payload := MapRequestToFHIR(req)
	payload.Set("id", id)
	return s.repo.Update(ctx, id, payload)
}

func (s *service) GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("Claim ID is required").Build()
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error) {
	return s.repo.Search(ctx, queryParams)
}

func (s *service) applyDefaults(req *ClaimRequest) {
	if req.OrganizationID == "" {
		req.OrganizationID = s.orgID
	}
	if req.ProviderID == "" {
		req.ProviderID = s.orgID
	}
}
