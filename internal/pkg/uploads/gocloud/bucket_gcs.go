package gocloud

import (
	"context"
	"fmt"
	"io/ioutil"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

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
	// GCSBlobConfig configures a gcs blob authentication method.
	GCSBlobConfig struct {
		GoogleAccessID     string
		PrivateKeyFilepath string
	}

	// GCSConfig configures a gcs based storage provider.
	GCSConfig struct {
		ServiceAccountKeyFilepath string
		Scopes                    []string
		BucketName                string
		BlobSettings              GCSBlobConfig
	}
)

func buildGCSBucket(ctx context.Context, cfg *GCSConfig) (*blob.Bucket, error) {
	creds, gcsCredsErr := gcp.DefaultCredentials(ctx)
	if gcsCredsErr != nil {
		return nil, fmt.Errorf("error constructing GCP credentials: %w", gcsCredsErr)
	}

	var (
		bucket *blob.Bucket
		err    error
	)

	if cfg.ServiceAccountKeyFilepath != "" {
		serviceAccountKeyBytes, accountKeyErr := ioutil.ReadFile(cfg.ServiceAccountKeyFilepath)
		if accountKeyErr != nil {
			return nil, fmt.Errorf("error reading service account key file: %w", accountKeyErr)
		}

		creds, gcsCredsErr = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes, cfg.Scopes...)
		if gcsCredsErr != nil {
			return nil, fmt.Errorf("error using service account key credentials: %w", gcsCredsErr)
		}
	}

	gcsClient, gcsClientErr := gcp.NewHTTPClient(nil, gcp.CredentialsTokenSource(creds))
	if gcsClientErr != nil {
		return nil, fmt.Errorf("error constructing GCP client: %w", gcsClientErr)
	}

	blobOpts := &gcsblob.Options{GoogleAccessID: cfg.BlobSettings.GoogleAccessID}

	if cfg.BlobSettings.PrivateKeyFilepath != "" {
		if blobOpts.PrivateKey, err = ioutil.ReadFile(cfg.ServiceAccountKeyFilepath); err != nil {
			return nil, fmt.Errorf("error reading private key file: %w", err)
		}
	}

	if bucket, err = gcsblob.OpenBucket(ctx, gcsClient, cfg.BucketName, blobOpts); err != nil {
		return nil, fmt.Errorf("error initializing filesystem bucket: %w", err)
	}

	return bucket, nil
}

// Validate validates the GCSConfig.
func (c *GCSConfig) Validate(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c, validation.Required),
	)
}
