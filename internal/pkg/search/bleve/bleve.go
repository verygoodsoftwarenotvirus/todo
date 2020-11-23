package bleve

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	bleve "github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/searcher"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	fuzziness = 2
	base      = 10
	bitSize   = 64

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

// NewBleveIndexManager instantiates a bleve index.
func NewBleveIndexManager(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
	var index bleve.Index

	preexistingIndex, openIndexErr := bleve.Open(string(path))
	if openIndexErr == nil {
		index = preexistingIndex
	}

	if errors.Is(openIndexErr, bleve.ErrorIndexPathDoesNotExist) {
		logger.WithValue("path", path).Debug("tried to open existing index, but didn't find it")

		var newIndexErr error

		switch name {
		case testingSearchIndexName:
			index, newIndexErr = bleve.New(string(path), bleve.NewIndexMapping())
			if newIndexErr != nil {
				logger.Error(newIndexErr, "failed to create new index")
				return nil, newIndexErr
			}
		case types.ItemsSearchIndexName:
			index, newIndexErr = bleve.New(string(path), buildItemMapping())
			if newIndexErr != nil {
				logger.Error(newIndexErr, "failed to create new index")
				return nil, newIndexErr
			}
		default:
			return nil, fmt.Errorf("invalid index name: %q", name)
		}
	} else if openIndexErr != nil {
		logger.Error(openIndexErr, "failed to open index")
		return nil, fmt.Errorf("failed to open index: %w", openIndexErr)
	}

	im := &bleveIndexManager{
		index:  index,
		logger: logger.WithName(fmt.Sprintf("%s_search", name)),
	}

	return im, nil
}

// Index implements our IndexManager interface.
func (sm *bleveIndexManager) Index(ctx context.Context, id uint64, value interface{}) error {
	_, span := tracing.StartSpan(ctx, "Index")
	defer span.End()

	sm.logger.WithValue("id", id).Debug("adding to index")

	return sm.index.Index(strconv.FormatUint(id, base), value)
}

// Search implements our IndexManager interface.
func (sm *bleveIndexManager) Search(ctx context.Context, query string, userID uint64) (ids []uint64, err error) {
	_, span := tracing.StartSpan(ctx, "Search")
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, query)
	sm.logger.WithValues(map[string]interface{}{
		"search_query": query,
		"user_id":      userID,
	}).Debug("performing search")

	q := bleve.NewFuzzyQuery(query)
	q.SetFuzziness(searcher.MaxFuzziness)

	searchResults, err := sm.index.SearchInContext(ctx, bleve.NewSearchRequest(q))
	if err != nil {
		sm.logger.Error(err, "performing search query")
		return nil, err
	}

	for _, result := range searchResults.Hits {
		x, err := strconv.ParseUint(result.ID, base, bitSize)
		if err != nil {
			// this should literally never happen
			return nil, fmt.Errorf("*gasp* impossible: %w", err)
		}

		ids = append(ids, x)
	}

	return ids, nil
}

// SearchForAdmin implements our IndexManager interface.
func (sm *bleveIndexManager) SearchForAdmin(ctx context.Context, query string) (ids []uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "SearchForAdmin")
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, query)
	logger := sm.logger.WithValue("search_query", query)
	logger.Debug("performing search for admin")

	q := bleve.NewFuzzyQuery(query)
	q.SetFuzziness(fuzziness)

	searchResults, err := sm.index.SearchInContext(ctx, bleve.NewSearchRequest(q))
	if err != nil {
		sm.logger.Error(err, "performing search query")
		return nil, err
	}

	for _, result := range searchResults.Hits {
		x, err := strconv.ParseUint(result.ID, base, bitSize)
		if err != nil {
			// this should literally never happen
			return nil, fmt.Errorf("*gasp* impossible: %w", err)
		}

		ids = append(ids, x)
	}

	return ids, nil
}

// Delete implements our IndexManager interface.
func (sm *bleveIndexManager) Delete(ctx context.Context, id uint64) error {
	_, span := tracing.StartSpan(ctx, "Delete")
	defer span.End()

	if err := sm.index.Delete(strconv.FormatUint(id, base)); err != nil {
		sm.logger.Error(err, "removing from index")
		return err
	}

	sm.logger.WithValue("id", id).Debug("removed from index")

	return nil
}
