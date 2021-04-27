package search

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

type (
	// IndexPath is a type alias for dependency injection's sake.
	IndexPath string

	// IndexName is a type alias for dependency injection's sake.
	IndexName string

	// IndexManager is our wrapper interface for a text search index.
	IndexManager interface {
		Index(ctx context.Context, id uint64, value interface{}) error
		Search(ctx context.Context, query string, accountID uint64) (ids []uint64, err error)
		SearchForAdmin(ctx context.Context, query string) (ids []uint64, err error)
		Delete(ctx context.Context, id uint64) (err error)
	}

	// IndexManagerProvider is a function that provides an IndexManager for a given index.
	IndexManagerProvider func(path IndexPath, name IndexName, logger logging.Logger) (IndexManager, error)
)
