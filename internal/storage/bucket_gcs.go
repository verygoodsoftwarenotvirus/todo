package storage

import (
	"context"
	"fmt"
	"os"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

const (
	// GCSProvider indicates we'd like to use the gcs adapter for blob.
	GCSProvider = "gcs"
)

type (
	// GCSBlobConfig configures a gcs blob passwords method.
	GCSBlobConfig struct {
		GoogleAccessID string `json:"google_access_id" mapstructure:"google_access_id" toml:"google_access_id,omitempty"`
	}

	// GCSConfig configures a gcs based storage provider.
	GCSConfig struct {
		BlobSettings              GCSBlobConfig `json:"blob_settings" mapstructure:"blob_settings" toml:"blob_settings,omitempty"`
		ServiceAccountKeyFilepath string        `json:"service_account_key_filepath" mapstructure:"service_account_key_filepath" toml:"service_account_key_filepath,omitempty"`
		BucketName                string        `json:"bucket_name" mapstructure:"bucket_name" toml:"bucket_name,omitempty"`
		Scopes                    []string      `json:"scopes" mapstructure:"scopes" toml:"scopes,omitempty"`
	}
)

func buildGCSBucket(ctx context.Context, cfg *GCSConfig) (*blob.Bucket, error) {
	var (
		creds  *google.Credentials
		bucket *blob.Bucket
	)

	if cfg.ServiceAccountKeyFilepath != "" {
		serviceAccountKeyBytes, err := os.ReadFile(cfg.ServiceAccountKeyFilepath)
		if err != nil {
			return nil, fmt.Errorf("reading service account key file: %w", err)
		}

		if creds, err = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes, cfg.Scopes...); err != nil {
			return nil, fmt.Errorf("using service account key credentials: %w", err)
		}
	} else {
		var err error
		if creds, err = gcp.DefaultCredentials(ctx); err != nil {
			return nil, fmt.Errorf("constructing GCPKMS credentials: %w", err)
		}
	}

	gcsClient, gcsClientErr := gcp.NewHTTPClient(nil, gcp.CredentialsTokenSource(creds))
	if gcsClientErr != nil {
		return nil, fmt.Errorf("constructing GCPKMS client: %w", gcsClientErr)
	}

	blobOpts := &gcsblob.Options{GoogleAccessID: cfg.BlobSettings.GoogleAccessID}

	bucket, err := gcsblob.OpenBucket(ctx, gcsClient, cfg.BucketName, blobOpts)
	if err != nil {
		return nil, fmt.Errorf("initializing filesystem bucket: %w", err)
	}

	return bucket, nil
}

var _ validation.ValidatableWithContext = (*GCSConfig)(nil)

// ValidateWithContext validates the GCSConfig.
func (c *GCSConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.BucketName, validation.Required),
	)
}
