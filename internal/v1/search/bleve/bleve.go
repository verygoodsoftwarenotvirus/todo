package bleve

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	bleve "github.com/blevesearch/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	base    = 10
	bitSize = 64

	// testingSearchIndexName is an index name that is only valid for testing's sake.
	testingSearchIndexName search.IndexName = "testing"
)

var _ search.IndexManager = (*bleveIndexManager)(nil)

type (
	bleveIndexManager struct {
		index  bleve.Index
		logger logging.Logger
	}
)

// NewBleveIndexManager instantiates a bleve index
func NewBleveIndexManager(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
	var index bleve.Index

	preexistingIndex, openIndexErr := bleve.Open(string(path))
	switch openIndexErr {
	case nil:
		index = preexistingIndex
	case bleve.ErrorIndexPathDoesNotExist:
		logger.WithValue("path", path).Debug("tried to open existing index, but didn't find it")
		var newIndexErr error

		switch name {
		case testingSearchIndexName:
			index, newIndexErr = bleve.New(string(path), bleve.NewIndexMapping())
			if newIndexErr != nil {
				logger.Error(newIndexErr, "failed to create new index")
				return nil, newIndexErr
			}
		case models.ItemsSearchIndexName:
			index, newIndexErr = bleve.New(string(path), buildItemMapping())
			if newIndexErr != nil {
				logger.Error(newIndexErr, "failed to create new index")
				return nil, newIndexErr
			}
		default:
			return nil, fmt.Errorf("invalid index name: %q", name)
		}
	default:
		logger.Error(openIndexErr, "failed to open index")
		return nil, openIndexErr
	}

	im := &bleveIndexManager{
		index:  index,
		logger: logger.WithName(fmt.Sprintf("%s_search", name)),
	}

	return im, nil
}

// Index implements our IndexManager interface
func (sm *bleveIndexManager) Index(ctx context.Context, id uint64, value interface{}) error {
	_, span := tracing.StartSpan(ctx, "Index")
	defer span.End()

	sm.logger.WithValue("id", id).Debug("adding to index")
	return sm.index.Index(strconv.FormatUint(id, base), value)
}

// Search implements our IndexManager interface
func (sm *bleveIndexManager) Search(ctx context.Context, query string, userID uint64) (ids []uint64, err error) {
	_, span := tracing.StartSpan(ctx, "Search")
	defer span.End()

	query = ensureQueryIsRestrictedToUser(query, userID)
	tracing.AttachSearchQueryToSpan(span, query)
	sm.logger.WithValues(map[string]interface{}{
		"search_query": query,
		"user_id":      userID,
	}).Debug("performing search")

	searchRequest := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	searchResults, err := sm.index.SearchInContext(ctx, searchRequest)
	if err != nil {
		sm.logger.Error(err, "performing search query")
		return nil, err
	}

	out := []uint64{}
	for _, result := range searchResults.Hits {
		x, err := strconv.ParseUint(result.ID, base, bitSize)
		if err != nil {
			// this should literally never happen
			return nil, err
		}
		out = append(out, x)
	}

	return out, nil
}

// Delete implements our IndexManager interface
func (sm *bleveIndexManager) Delete(ctx context.Context, id uint64) error {
	_, span := tracing.StartSpan(ctx, "Delete")
	defer span.End()

	sm.logger.WithValue("id", id).Debug("removing from index")
	return sm.index.Delete(strconv.FormatUint(id, base))
}
