package encounter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"service/internal/infrastructure/database"
	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, payload interface{}) (*satusehat.FHIRResponse, error)
	Update(ctx context.Context, id string, payload interface{}) (*satusehat.FHIRResponse, error)
	Patch(ctx context.Context, id string, payload EncounterPatchRequest) (*satusehat.FHIRResponse, error)
	GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error)
	Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error)
	SearchPatientByNIK(ctx context.Context, nik string) (*satusehat.FHIRResponse, error)
	SearchPractitionerByNIK(ctx context.Context, nik string) (*satusehat.FHIRResponse, error)
}

type repository struct {
	client satusehat.SatuSehatClient
	db     database.Service
}

func NewRepository(client satusehat.SatuSehatClient, db database.Service) Repository {
	return &repository{client: client, db: db}
}

func (r *repository) executeRequest(ctx context.Context, method, endpoint string, payload interface{}) (*satusehat.FHIRResponse, error) {
	resp, err := r.client.DoRequest(ctx, method, endpoint, payload)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to execute request to SatuSehat").Cause(err).Build()
	}
	return r.parseAndProcessResponse(resp)
}

func (r *repository) parseAndProcessResponse(data []byte) (*satusehat.FHIRResponse, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.InternalError().Message("Failed to parse SatuSehat response").Cause(err).Metadata("raw_response", string(data)).Build()
	}
	if resourceType, ok := result["resourceType"].(string); ok && resourceType == "OperationOutcome" {
		return nil, errors.ParseSatuSehatError(result)
	}
	var resourceID string
	if id, ok := result["id"].(string); ok {
		resourceID = id
	}
	return &satusehat.FHIRResponse{ID: resourceID, FullResponse: result, RawResponse: data}, nil
}

func (r *repository) Create(ctx context.Context, payload interface{}) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "POST", "/Encounter", payload)
}

func (r *repository) Update(ctx context.Context, id string, payload interface{}) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "PUT", fmt.Sprintf("/Encounter/%s", id), payload)
}

func (r *repository) Patch(ctx context.Context, id string, payload EncounterPatchRequest) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "PATCH", fmt.Sprintf("/Encounter/%s", id), payload)
}

func (r *repository) GetByID(ctx context.Context, id string) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/Encounter/%s", id), nil)
}

func (r *repository) Search(ctx context.Context, queryParams url.Values) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/Encounter?%s", queryParams.Encode()), nil)
}

func (r *repository) SearchPatientByNIK(ctx context.Context, nik string) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/Patient?identifier=https://fhir.kemkes.go.id/id/nik|%s", nik), nil)
}

func (r *repository) SearchPractitionerByNIK(ctx context.Context, nik string) (*satusehat.FHIRResponse, error) {
	return r.executeRequest(ctx, "GET", fmt.Sprintf("/Practitioner?identifier=https://fhir.kemkes.go.id/id/nik|%s", nik), nil)
}
