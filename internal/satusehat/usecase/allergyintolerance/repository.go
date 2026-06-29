package allergyintolerance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, payload interface{}) (*satusehat.FHIRResponse, error)
	Update(ctx context.Context, id string, payload interface{}) (*satusehat.FHIRResponse, error)
	Patch(ctx context.Context, id string, req AllergyIntolerancePatchRequest) (*satusehat.FHIRResponse, error)
	GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error)
	Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error)
}
type repository struct{ client satusehat.SatuSehatClient }

func NewRepository(client satusehat.SatuSehatClient) Repository { return &repository{client: client} }

func (r *repository) executeRequest(ctx context.Context, method, endpoint string, req interface{}) (*satusehat.FHIRResponse, error) {
	resp, err := r.client.DoRequest(ctx, method, endpoint, req)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to execute request to SatuSehat").Cause(err).Build()
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, errors.InternalError().Message("Failed to parse SatuSehat response").Cause(err).Metadata("raw_response", string(resp)).Build()
	}

	if resourceType, ok := result["resourceType"].(string); ok && resourceType == "OperationOutcome" {
		return nil, errors.ParseSatuSehatError(result)
	}

	var resourceID string
	if id, ok := result["id"].(string); ok {
		resourceID = id
	}
	return &satusehat.FHIRResponse{ID: resourceID, FullResponse: result, RawResponse: resp}, nil
}
func (r *repository) Create(ctx context.Context, payload interface{}) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "POST", "/AllergyIntolerance", payload)
}
func (r *repository) Update(ctx context.Context, id string, payload interface{}) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "PUT", fmt.Sprintf("/AllergyIntolerance/%s", id), payload)
}
func (r *repository) Patch(ctx context.Context, id string, req AllergyIntolerancePatchRequest) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "PATCH", fmt.Sprintf("/AllergyIntolerance/%s", id), req)
}
func (r *repository) GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/AllergyIntolerance/%s", id), nil)
}
func (r *repository) Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/AllergyIntolerance?%s", queryParams.Encode()), nil)
}
