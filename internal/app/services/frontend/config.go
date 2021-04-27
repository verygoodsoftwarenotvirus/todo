package frontend

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config describes the settings pertinent to the frontend.
type Config struct {
	StaticFilesDirectory string `json:"static_files_directory" mapstructure:"static_files_directory" toml:"static_files_directory,omitempty"`
	Debug                bool   `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	LogStaticFiles       bool   `json:"log_static_files" mapstructure:"log_static_files" toml:"log_static_files,omitempty"`
	CacheStaticFiles     bool   `json:"cache_static_files" mapstructure:"cache_static_files" toml:"cache_static_files,omitempty"`
}

// ValidateWithContext validates a Config struct.
func (s Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.StaticFilesDirectory, validation.Required),
	)
}
