package minio

import (
	"errors"
	"os"
	"strings"

	logs "github.com/karincake/apem/loggero"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var O MinioCfg = MinioCfg{}
var I *minio.Client

type MinioCfg struct {
	Endpoint   string   `yaml:"endpoint" env:"MINIO_ENDPOINT"`
	Region     string   `yaml:"region" env:"MINIO_REGION"`
	AccessKey  string   `yaml:"accessKey" env:"MINIO_ACCESSKEY"`
	SecretKey  string   `yaml:"secretKey" env:"MINIO_SECRETKEY"`
	UseSsl     bool     `yaml:"useSsl" env:"MINIO_USESSL"`
	BucketName []string `yaml:"bucketName" env:"MINIO_BUCKETNAME"`
}

type ResponsePostPolicy struct {
	Url      string            `json:"url"`
	FormData map[string]string `json:"form-data"`
}

func (c MinioCfg) GetRegion() string {
	return c.Region
}

func (c MinioCfg) GetEndpoint() string {
	return c.Endpoint
}

func (c MinioCfg) GetUseSsl() bool {
	return c.UseSsl
}

func (c MinioCfg) GetBucketName() []string {
	return c.BucketName
}

// connect db
func Connect() {
	// Memastikan data konfigurasi dari file .env teraplikasikan
	if O.Endpoint == "" && os.Getenv("MINIO_ENDPOINT") != "" {
		O.Endpoint = os.Getenv("MINIO_ENDPOINT")
	}
	if O.AccessKey == "" && os.Getenv("MINIO_ACCESSKEY") != "" {
		O.AccessKey = os.Getenv("MINIO_ACCESSKEY")
	}
	if O.SecretKey == "" && os.Getenv("MINIO_SECRETKEY") != "" {
		O.SecretKey = os.Getenv("MINIO_SECRETKEY")
	}
	if O.Region == "" && os.Getenv("MINIO_REGION") != "" {
		O.Region = os.Getenv("MINIO_REGION")
	}
	if os.Getenv("MINIO_USESSL") == "true" {
		O.UseSsl = true
	}

	err := NewClient(&O)
	if err != nil {
		panic("minio client initialization failed: " + err.Error())
	}
	if I == nil {
		panic("minio client is nil")
	}
	logs.I.Println("Instantiation for object storage service using Minio, status: DONE!!")
}

func NewClient(cfg *MinioCfg) error {
	// Initialize minio client object.
	endpoint := cfg.Endpoint

	// minio-go SDK requires endpoint without http:// or https:// scheme
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	if endpoint == "" {
		return errors.New("config minio endpoint empty")
	}
	accessKey := cfg.AccessKey
	if accessKey == "" {
		return errors.New("config minio access key empty")
	}
	secretKey := cfg.SecretKey
	if secretKey == "" {
		return errors.New("config minio secret key empty")
	}
	useSSL := cfg.UseSsl
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}
	I = minioClient
	return nil
}
