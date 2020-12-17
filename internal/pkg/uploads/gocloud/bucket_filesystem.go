package gocloud

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// FilesystemProvider indicates we'd like to use the filesystem adapter for blob.
	FilesystemProvider = "filesystem"
)

type (
	// FilesystemConfig configures a filesystem-based storage provider.
	FilesystemConfig struct {
		RootDirectory string
	}
)

// Validate validates the FilesystemConfig.
func (c *FilesystemConfig) Validate(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.RootDirectory, validation.Required),
	)
}
