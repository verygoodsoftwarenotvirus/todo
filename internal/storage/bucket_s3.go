package storage

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// S3Provider indicates we'd like to use the s3 adapter for blob.
	S3Provider = "s3"
)

type (
	// S3Config configures an S3-based storage provider.
	S3Config struct {
		_ struct{}

		BucketName string `json:"bucket_name" mapstructure:"bucket_name" toml:"bucket_name,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*S3Config)(nil)

// ValidateWithContext validates the S3Config.
func (c *S3Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.BucketName, validation.Required),
	)
}
