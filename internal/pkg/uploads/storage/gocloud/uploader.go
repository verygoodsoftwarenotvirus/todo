package gocloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"

	"github.com/aws/aws-sdk-go/aws/session"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"
)

const (
	// MemoryProvider indicates we'd like to use the memory adapter for blob.
	MemoryProvider = "memory"
)

var (
	// ErrInvalidConfiguration denotes that the configuration is invalid.
	ErrInvalidConfiguration = errors.New("configuration invalid")
)

type (
	// Uploader implements our UploadManager struct.
	Uploader struct {
		bucket *blob.Bucket
		logger logging.Logger
		tracer tracing.Tracer
	}

	// UploaderConfig configures our UploadManager.
	UploaderConfig struct {
		Provider         string            `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		Name             string            `json:"name" mapstructure:"name" toml:"name,omitempty"`
		AzureConfig      *AzureConfig      `json:"azure" mapstructure:"azure" toml:"azure,omitempty"`
		GCSConfig        *GCSConfig        `json:"gcs" mapstructure:"gcs" toml:"gcs,omitempty"`
		S3Config         *S3Config         `json:"s3" mapstructure:"s3" toml:"s3,omitempty"`
		FilesystemConfig *FilesystemConfig `json:"filesystem" mapstructure:"filesystem" toml:"filesystem,omitempty"`
	}
)

// Validate validates the UploaderConfig.
func (c *UploaderConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Provider, validation.In(AzureProvider, GCSProvider, S3Provider, FilesystemProvider, MemoryProvider)),
		validation.Field(&c.AzureConfig, validation.When(c.Provider == AzureProvider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.GCSConfig, validation.When(c.Provider == GCSProvider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.S3Config, validation.When(c.Provider == S3Provider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.FilesystemConfig, validation.When(c.Provider == FilesystemProvider, validation.Required).Else(validation.Nil)),
	)
}

// NewUploadManager provides a new uploads.UploadManager.
func NewUploadManager(ctx context.Context, logger logging.Logger, cfg UploaderConfig) (uploads.UploadManager, error) {
	serviceName := fmt.Sprintf("%s_uploader", cfg.Name)
	u := &Uploader{
		logger: logger.WithName(serviceName),
		tracer: tracing.NewTracer(serviceName),
	}

	if err := cfg.Validate(ctx); err != nil {
		return nil, fmt.Errorf("upload manager provided invalid config: %w", err)
	}

	var err error

	if u.bucket, err = selectBucket(ctx, logger, cfg); err != nil {
		return nil, fmt.Errorf("error initializing bucket: %w", err)
	}

	return u, nil
}

func selectBucket(ctx context.Context, logger logging.Logger, cfg UploaderConfig) (bucket *blob.Bucket, err error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case AzureProvider:
		if cfg.AzureConfig == nil {
			return nil, ErrInvalidConfiguration
		}

		if bucket, err = provideAzureBucket(ctx, cfg.AzureConfig, logger); err != nil {
			return nil, fmt.Errorf("error initializing azure bucket: %w", err)
		}
	case GCSProvider:
		if cfg.GCSConfig == nil {
			return nil, ErrInvalidConfiguration
		}

		if bucket, err = buildGCSBucket(ctx, cfg.GCSConfig); err != nil {
			return nil, fmt.Errorf("error initializing gcs bucket: %w", err)
		}
	case S3Provider:
		if cfg.S3Config == nil {
			return nil, ErrInvalidConfiguration
		}

		if bucket, err = s3blob.OpenBucket(ctx, session.Must(session.NewSession()), cfg.S3Config.BucketName, &s3blob.Options{
			UseLegacyList: false,
		}); err != nil {
			return nil, fmt.Errorf("error initializing s3 bucket: %w", err)
		}
	case MemoryProvider:
		bucket = memblob.OpenBucket(&memblob.Options{})
	default:
		if cfg.FilesystemConfig == nil {
			return nil, ErrInvalidConfiguration
		}

		if bucket, err = fileblob.OpenBucket(cfg.FilesystemConfig.RootDirectory, &fileblob.Options{
			URLSigner: nil,
			CreateDir: true,
		}); err != nil {
			return nil, fmt.Errorf("error initializing filesystem bucket: %w", err)
		}
	}

	bucket = blob.PrefixedBucket(bucket, cfg.Name)

	return bucket, err
}

// SaveFile saves a file to the blob.
func (u *Uploader) SaveFile(ctx context.Context, path string, content []byte) error {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	if err := u.bucket.WriteAll(ctx, path, content, nil); err != nil {
		return fmt.Errorf("error writing file content: %w", err)
	}

	return nil
}

// ReadFile reads a file from the blob.
func (u *Uploader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	r, err := u.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching file: %w", err)
	}

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return bytes, nil
}
