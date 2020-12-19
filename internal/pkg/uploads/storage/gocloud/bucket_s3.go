package gocloud

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
		BucketName string
	}
)

// Validate validates the S3Config.
func (c *S3Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.BucketName, validation.Required),
	)
}
