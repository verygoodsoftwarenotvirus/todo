package config

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

// SearchSettings contains settings regarding search indices.
type SearchSettings struct {
	// ItemsIndexPath indicates where our items search index files should go.
	ItemsIndexPath search.IndexPath `json:"items_index_path" mapstructure:"items_index_path" toml:"items_index_path,omitempty"`
}
