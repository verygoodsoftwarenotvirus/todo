package config

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// SearchSettings contains settings regarding search indices.
type SearchSettings struct {
	// ItemsIndexPath indicates where our items search index files should go.
	ItemsIndexPath search.IndexPath `json:"items_index_path" mapstructure:"items_index_path" toml:"items_index_path,omitempty"`
}

// Validate validates a SearchSettings struct.
func (s SearchSettings) Validate(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.ItemsIndexPath, validation.Required),
	)
}
