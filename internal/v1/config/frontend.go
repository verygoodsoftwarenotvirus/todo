package config

// FrontendSettings describes the settings pertinent to the frontend.
type FrontendSettings struct {
	// StaticFilesDirectory indicates which directory contains our static files for the frontend (i.e. CSS/JS/HTML files)
	StaticFilesDirectory string `json:"static_files_directory" mapstructure:"static_files_directory" toml:"static_files_directory,omitempty"`
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// CacheStaticFiles indicates whether or not to load the static files directory into memory via afero's MemMapFs.
	CacheStaticFiles bool `json:"cache_static_files" mapstructure:"cache_static_files" toml:"cache_static_files,omitempty"`
}
