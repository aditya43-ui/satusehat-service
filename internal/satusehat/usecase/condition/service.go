package condition

import (
	"context"
	"net/url"
	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, req ConditionRequest) (*satusehat.FHIRResponse, error)
	Update(ctx context.Context, id string, req ConditionRequest) (*satusehat.FHIRResponse, error)
	Patch(ctx context.Context, id string, req ConditionPatchRequest) (*satusehat.FHIRResponse, error)
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

func (s *service) Create(ctx context.Context, req ConditionRequest) (*satusehat.FHIRResponse, error) {
	if req.OrganizationID == "" {
		req.OrganizationID = s.orgID
	}
	return s.repo.Create(ctx, MapRequestToFHIR(req))
}

func (s *service) Update(ctx context.Context, id string, req ConditionRequest) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("Condition ID is required").Build()
	}
	if req.OrganizationID == "" {
		req.OrganizationID = s.orgID
	}
	fhirPayload := MapRequestToFHIR(req)
	fhirPayload.Set("id", id)

	return s.repo.Update(ctx, id, fhirPayload)
}

func (s *service) Patch(ctx context.Context, id string, req ConditionPatchRequest) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("Condition ID is required").Build()
	}
	if len(req) == 0 {
		return nil, errors.NewValidationError().Message("Patch payload cannot be empty").Build()
	}
	return s.repo.Patch(ctx, id, req)
}

func (s *service) GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error) {
	if id == "" {
		return nil, errors.NewValidationError().Message("Condition ID is required").Build()
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error) {
	return s.repo.Search(ctx, queryParams)
}
