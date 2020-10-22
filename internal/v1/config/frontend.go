package config

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"
)

// FrontendSettings describes the settings pertinent to the frontend.
type FrontendSettings struct {
	// StaticFilesDirectory indicates which directory contains our static files for the frontend (i.e. CSS/JS/HTML files)
	StaticFilesDirectory string `json:"static_files_directory" mapstructure:"static_files_directory" toml:"static_files_directory,omitempty"`
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// CacheStaticFiles indicates whether or not to load the static files directory into memory via afero's MemMapFs.
	CacheStaticFiles bool `json:"cache_static_files" mapstructure:"cache_static_files" toml:"cache_static_files,omitempty"`
}

// SearchSettings contains settings regarding search indices.
type SearchSettings struct {
	// ItemsIndexPath indicates where our items search index files should go.
	ItemsIndexPath search.IndexPath `json:"items_index_path" mapstructure:"items_index_path" toml:"items_index_path,omitempty"`
}
