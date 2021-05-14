package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/aws/aws-sdk-go/aws/session"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
)

const (
	// MemoryProvider indicates we'd like to use the memory adapter for blob.
	MemoryProvider = "memory"
)

var (
	// ErrNilConfig denotes that the provided configuration is nil.
	ErrNilConfig = errors.New("nil config provided")

	// ErrInvalidConfiguration denotes that the provided configuration is invalid.
	ErrInvalidConfiguration = errors.New("configuration invalid")

	// ErrBucketIsUnavailable denotes that the requested bucket is invalid.
	ErrBucketIsUnavailable = errors.New("bucket is unavailable")
)

type (
	// Uploader implements our UploadManager struct.
	Uploader struct {
		bucket          *blob.Bucket
		logger          logging.Logger
		tracer          tracing.Tracer
		filenameFetcher func(req *http.Request) string
	}

	// Config configures our UploadManager.
	Config struct {
		FilesystemConfig  *FilesystemConfig `json:"filesystem" mapstructure:"filesystem" toml:"filesystem,omitempty"`
		AzureConfig       *AzureConfig      `json:"azure" mapstructure:"azure" toml:"azure,omitempty"`
		GCSConfig         *GCSConfig        `json:"gcs" mapstructure:"gcs" toml:"gcs,omitempty"`
		S3Config          *S3Config         `json:"s3" mapstructure:"s3" toml:"s3,omitempty"`
		BucketName        string            `json:"bucket_name" mapstructure:"bucket_name" toml:"bucket_name,omitempty"`
		UploadFilenameKey string            `json:"upload_filename_key" mapstructure:"upload_filename_key" toml:"upload_filename_key,omitempty"`
		Provider          string            `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the Config.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.BucketName, validation.Required),
		validation.Field(&c.Provider, validation.In(AzureProvider, GCSProvider, S3Provider, FilesystemProvider, MemoryProvider)),
		validation.Field(&c.AzureConfig, validation.When(c.Provider == AzureProvider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.GCSConfig, validation.When(c.Provider == GCSProvider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.S3Config, validation.When(c.Provider == S3Provider, validation.Required).Else(validation.Nil)),
		validation.Field(&c.FilesystemConfig, validation.When(c.Provider == FilesystemProvider, validation.Required).Else(validation.Nil)),
	)
}

// NewUploadManager provides a new uploads.UploadManager.
func NewUploadManager(ctx context.Context, logger logging.Logger, cfg *Config, routeParamManager routing.RouteParamManager) (*Uploader, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	serviceName := fmt.Sprintf("%s_uploader", cfg.BucketName)
	u := &Uploader{
		logger:          logging.EnsureLogger(logger).WithName(serviceName),
		tracer:          tracing.NewTracer(serviceName),
		filenameFetcher: routeParamManager.BuildRouteParamStringIDFetcher(cfg.UploadFilenameKey),
	}

	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("upload manager provided invalid config: %w", err)
	}

	if err := u.selectBucket(ctx, cfg); err != nil {
		return nil, fmt.Errorf("initializing bucket: %w", err)
	}

	if available, err := u.bucket.IsAccessible(ctx); err != nil {
		return nil, fmt.Errorf("verifying bucket accessibility: %w", err)
	} else if !available {
		return nil, ErrBucketIsUnavailable
	}

	return u, nil
}

func (u *Uploader) selectBucket(ctx context.Context, cfg *Config) (err error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case AzureProvider:
		if cfg.AzureConfig == nil {
			return ErrNilConfig
		}

		if u.bucket, err = provideAzureBucket(ctx, cfg.AzureConfig, u.logger); err != nil {
			return fmt.Errorf("initializing azure bucket: %w", err)
		}
	case GCSProvider:
		if cfg.GCSConfig == nil {
			return ErrNilConfig
		}

		if u.bucket, err = buildGCSBucket(ctx, cfg.GCSConfig); err != nil {
			return fmt.Errorf("initializing gcs bucket: %w", err)
		}
	case S3Provider:
		if cfg.S3Config == nil {
			return ErrNilConfig
		}

		if u.bucket, err = s3blob.OpenBucket(ctx, session.Must(session.NewSession()), cfg.S3Config.BucketName, &s3blob.Options{
			UseLegacyList: false,
		}); err != nil {
			return fmt.Errorf("initializing s3 bucket: %w", err)
		}
	case MemoryProvider:
		u.bucket = memblob.OpenBucket(&memblob.Options{})
	default:
		if cfg.FilesystemConfig == nil {
			return ErrNilConfig
		}

		if u.bucket, err = fileblob.OpenBucket(cfg.FilesystemConfig.RootDirectory, &fileblob.Options{
			URLSigner: nil,
			CreateDir: true,
		}); err != nil {
			return fmt.Errorf("initializing filesystem bucket: %w", err)
		}
	}

	bn := cfg.BucketName
	if !strings.HasSuffix(bn, "_") {
		bn = fmt.Sprintf("%s_", cfg.BucketName)
	}

	u.bucket = blob.PrefixedBucket(u.bucket, bn)

	return err
}
