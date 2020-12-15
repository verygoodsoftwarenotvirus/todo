package config

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// UploadSettings contains settings regarding search indices.
type UploadSettings struct {
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// RootPath is the location where all upload subdirectories belong
	RootPath uploads.RootUploadDirectory `json:"root_upload_directory" mapstructure:"root_upload_directory" toml:"root_upload_directory,omitempty"`
}

// Validate validates an UploadSettings struct.
func (s UploadSettings) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.RootPath, validation.Required),
	)
}
