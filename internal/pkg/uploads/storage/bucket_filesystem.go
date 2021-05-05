package storage

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// FilesystemProvider indicates we'd like to use the filesystem adapter for blob.
	FilesystemProvider = "filesystem"
)

type (
	// FilesystemConfig configures a filesystem-based storage provider.
	FilesystemConfig struct {
		RootDirectory string `json:"root_directory" mapstructure:"root_directory" toml:"root_directory,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*FilesystemConfig)(nil)

// ValidateWithContext validates the FilesystemConfig.
func (c *FilesystemConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.RootDirectory, validation.Required),
	)
}
