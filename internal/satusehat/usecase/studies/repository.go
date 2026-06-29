package studies

import (
	"context"
	"encoding/json"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
)

type Repository interface {
	UploadDICOM(ctx context.Context, dicomBytes []byte) (*satusehat.FHIRResponse, error)
}

type repository struct {
	client satusehat.SatuSehatClient
}

func NewRepository(client satusehat.SatuSehatClient) Repository {
	return &repository{client: client}
}

func (r *repository) UploadDICOM(ctx context.Context, dicomBytes []byte) (*satusehat.FHIRResponse, error) {
	respData, err := r.client.UploadDICOM(ctx, dicomBytes)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, errors.InternalError().Message("Gagal memparsing response DICOM dari SatuSehat").Cause(err).Build()
	}

	return &satusehat.FHIRResponse{
		FullResponse: result,
		RawResponse:  respData,
	}, nil
}
