package search

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

type (
	// IndexPath is a type alias for dependency injection's sake.
	IndexPath string

	// IndexName is a type alias for dependency injection's sake.
	IndexName string

	// IndexManager is our wrapper interface for a text search index.
	IndexManager interface {
		Index(ctx context.Context, id string, value interface{}) error
		Search(ctx context.Context, query, accountID string) (ids []string, err error)
		SearchForAdmin(ctx context.Context, query string) (ids []string, err error)
		Delete(ctx context.Context, id string) (err error)
	}

	// IndexManagerProvider is a function that provides an IndexManager for a given index.
	IndexManagerProvider func(path IndexPath, name IndexName, logger logging.Logger) (IndexManager, error)
)
