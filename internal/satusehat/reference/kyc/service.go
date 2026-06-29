package kyc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"service/internal/infrastructure/config"
	"service/internal/interfaces/satusehat"
	"service/pkg/crypto"
	"service/pkg/errors"
	"service/pkg/logger"
)

type Service interface {
	GenerateURL(ctx context.Context, req GenerateURLRequest) (*satusehat.FHIRResponse, error)
	HandleCallback(ctx context.Context, encryptedPayload string, token string) error
}

type service struct {
	repo Repository
	cfg  config.SatuSehatConfig
}

func NewService(repo Repository, cfg config.SatuSehatConfig) Service {
	return &service{repo: repo, cfg: cfg}
}

func (s *service) GenerateURL(ctx context.Context, req GenerateURLRequest) (*satusehat.FHIRResponse, error) {
	// Gunakan public key statis dari konfigurasi (disimpan dalam format Base64 di .env)
	pubKeyBytes, err := base64.StdEncoding.DecodeString(s.cfg.KYCPublicKeyB64)
	if err != nil {
		return nil, errors.InternalError().Message("Gagal mendecode Public Key dari konfigurasi").Cause(err).Build()
	}
	req.PublicKey = string(pubKeyBytes)

	// Decode private key untuk mendekripsi balasan generate URL dari Kemenkes
	privKeyBytes, err := base64.StdEncoding.DecodeString(s.cfg.KYCPrivateKeyB64)
	if err != nil {
		return nil, errors.InternalError().Message("Gagal mendecode Private Key dari konfigurasi").Cause(err).Build()
	}
	privKeyPEM := string(privKeyBytes)

	// Lempar privKeyPEM ke repo agar Kemenkes response dapat didekripsi
	return s.repo.GenerateURL(ctx, req, privKeyPEM)
}

func (s *service) HandleCallback(ctx context.Context, encryptedPayload string, token string) error {
	// Validasi Security Token dari Kemenkes
	expectedToken := s.cfg.WebhookSecret
	if expectedToken == "" {
		logger.Default().Warn("SATUSEHAT_WEBHOOK_SECRET belum dikonfigurasi. Validasi token webhook dilewati.")
	} else if token != expectedToken {
		logger.Default().Warn("Webhook KYC gagal divalidasi: Token tidak cocok", logger.String("received_token", token))
		return errors.UnauthorizedError().Message("Invalid security webhook token").Build()
	}

	// 1. Dapatkan Private Key statis Anda dari config/env var
	privKeyBytes, err := base64.StdEncoding.DecodeString(s.cfg.KYCPrivateKeyB64)
	if err != nil {
		return errors.InternalError().Message("Gagal mendecode Private Key dari konfigurasi").Cause(err).Build()
	}
	privKeyPEM := string(privKeyBytes)

	// 2. Dekripsi payload yang masuk
	decryptedBytes, err := crypto.DecryptSatuSehatPayload(encryptedPayload, privKeyPEM)
	if err != nil {
		return errors.InternalError().Message("Gagal mendekripsi payload webhook dari SatuSehat").Cause(err).Build()
	}

	// 3. Unmarshal ke bentuk Struct
	var payload CallbackRequest
	if err := json.Unmarshal(decryptedBytes, &payload); err != nil {
		return errors.InternalError().Message("Gagal memparsing JSON hasil dekripsi").Cause(err).Build()
	}

	// TODO: Tambahkan logika update ke database lokal (misal: tabel user/agent) berdasarkan hasil KYC.
	// Saat ini, kita log payload yang diterima dari Kemenkes.
	logger.Default().Info("Menerima Webhook KYC dari SatuSehat", logger.Any("payload", payload))
	return nil
}
