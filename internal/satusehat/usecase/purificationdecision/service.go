package purificationdecision

import (
	"context"
	"net/url"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, req PurificationDecisionRequest) (*satusehat.FHIRResponse, error)
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

func (s *service) Create(ctx context.Context, req PurificationDecisionRequest) (*satusehat.FHIRResponse, error) {
	if req.OrganizationID == "" {
		req.OrganizationID = s.orgID
	}
	return s.repo.Create(ctx, MapRequestToFHIR(req))
}

func (s *service) GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("PurificationDecision ID is required").Build()
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error) {
	return s.repo.Search(ctx, queryParams)
}
