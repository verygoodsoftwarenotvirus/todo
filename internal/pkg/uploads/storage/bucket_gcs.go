package storage

import (
	"context"
	"fmt"
	"io/ioutil"

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
		GoogleAccessID     string `json:"google_access_id" mapstructure:"google_access_id" toml:"google_access_id,omitempty"`
		PrivateKeyFilepath string `json:"private_key_filepath" mapstructure:"private_key_filepath" toml:"private_key_filepath,omitempty"`
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
	creds, gcsCredsErr := gcp.DefaultCredentials(ctx)
	if gcsCredsErr != nil {
		return nil, fmt.Errorf("constructing GCP credentials: %w", gcsCredsErr)
	}

	var (
		bucket *blob.Bucket
		err    error
	)

	if cfg.ServiceAccountKeyFilepath != "" {
		serviceAccountKeyBytes, accountKeyErr := ioutil.ReadFile(cfg.ServiceAccountKeyFilepath)
		if accountKeyErr != nil {
			return nil, fmt.Errorf("reading service account key file: %w", accountKeyErr)
		}

		creds, gcsCredsErr = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes, cfg.Scopes...)
		if gcsCredsErr != nil {
			return nil, fmt.Errorf("using service account key credentials: %w", gcsCredsErr)
		}
	}

	gcsClient, gcsClientErr := gcp.NewHTTPClient(nil, gcp.CredentialsTokenSource(creds))
	if gcsClientErr != nil {
		return nil, fmt.Errorf("constructing GCP client: %w", gcsClientErr)
	}

	blobOpts := &gcsblob.Options{GoogleAccessID: cfg.BlobSettings.GoogleAccessID}

	if cfg.BlobSettings.PrivateKeyFilepath != "" {
		if blobOpts.PrivateKey, err = ioutil.ReadFile(cfg.ServiceAccountKeyFilepath); err != nil {
			return nil, fmt.Errorf("reading private key file: %w", err)
		}
	}

	if bucket, err = gcsblob.OpenBucket(ctx, gcsClient, cfg.BucketName, blobOpts); err != nil {
		return nil, fmt.Errorf("initializing filesystem bucket: %w", err)
	}

	return bucket, nil
}

// ValidateWithContext validates the GCSConfig.
func (c *GCSConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c, validation.Required),
	)
}
