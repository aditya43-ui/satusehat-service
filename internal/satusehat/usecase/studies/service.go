package studies

import (
	"context"

	"service/internal/interfaces/satusehat"
	"service/pkg/errors"
	"service/pkg/logger"
)

type Service interface {
	UploadDICOM(ctx context.Context, dicomBytes []byte) (*satusehat.FHIRResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) UploadDICOM(ctx context.Context, dicomBytes []byte) (*satusehat.FHIRResponse, error) {
	if len(dicomBytes) == 0 {
		return nil, errors.NewValidationError().Message("File DICOM kosong").Build()
	}

	logger.Default().Info("Mempersiapkan upload DICOM", logger.Int("file_size_bytes", len(dicomBytes)))

	resp, err := s.repo.UploadDICOM(ctx, dicomBytes)
	if err != nil {
		return nil, errors.InternalError().Message("Gagal meneruskan file DICOM ke SatuSehat").Cause(err).Build()
	}

	return resp, nil
}
