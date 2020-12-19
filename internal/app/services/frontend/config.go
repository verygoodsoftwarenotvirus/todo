package frontend

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config describes the settings pertinent to the frontend.
type Config struct {
	// StaticFilesDirectory indicates which directory contains our static files for the frontend (i.e. CSS/JS/HTML files)
	StaticFilesDirectory string `json:"static_files_directory" mapstructure:"static_files_directory" toml:"static_files_directory,omitempty"`
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// LogStaticFiles determines if we log static file requests.
	LogStaticFiles bool `json:"log_static_files" mapstructure:"log_static_files" toml:"log_static_files,omitempty"`
	// CacheStaticFiles indicates whether or not to load the static files directory into memory via afero's MemMapFs.
	CacheStaticFiles bool `json:"cache_static_files" mapstructure:"cache_static_files" toml:"cache_static_files,omitempty"`
}

// Validate validates a Config struct.
func (s Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.StaticFilesDirectory, validation.Required),
	)
}
