package bleve

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/searcher"
)

const (
	// testingSearchIndexName is an index name that is only valid for testing's sake.
	testingSearchIndexName search.IndexName = "example_index_name"
)

var (
	errInvalidIndexName = errors.New("invalid index name")
)

var _ search.IndexManager = (*bleveIndexManager)(nil)

type (
	bleveIndexManager struct {
		index  bleve.Index
		logger logging.Logger
		tracer tracing.Tracer
	}
)

// NewBleveIndexManager instantiates a bleve index.
func NewBleveIndexManager(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
	var index bleve.Index

	preexistingIndex, err := bleve.Open(string(path))
	if err == nil {
		index = preexistingIndex
	}

	if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) || errors.Is(err, bleve.ErrorIndexMetaMissing) {
		logger.WithValue("path", path).Debug("tried to open existing index, but didn't find it")

		switch name {
		case testingSearchIndexName:
			index, err = bleve.New(string(path), bleve.NewIndexMapping())
		case types.ItemsSearchIndexName:
			index, err = bleve.New(string(path), buildItemMapping())
		default:
			return nil, fmt.Errorf("opening %s index: %w", name, errInvalidIndexName)
		}

		if err != nil {
			logger.Error(err, "failed to create new index")
			return nil, err
		}
	}

	serviceName := fmt.Sprintf("%s_search", name)

	im := &bleveIndexManager{
		index:  index,
		logger: logging.EnsureLogger(logger).WithName(serviceName),
		tracer: tracing.NewTracer(serviceName),
	}

	return im, nil
}

// Index implements our IndexManager interface.
func (sm *bleveIndexManager) Index(ctx context.Context, id string, value interface{}) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	sm.logger.WithValue("id", id).Debug("adding to index")

	return sm.index.Index(id, value)
}

// search executes search queries.
func (sm *bleveIndexManager) search(ctx context.Context, query, accountID string, forServiceAdmin bool) (ids []string, err error) {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, query)
	logger := sm.logger.WithValue(keys.SearchQueryKey, query)

	if query == "" {
		return nil, search.ErrEmptyQueryProvided
	}

	if !forServiceAdmin && accountID != "" {
		logger = logger.WithValue(keys.AccountIDKey, accountID)
	}

	q := bleve.NewFuzzyQuery(query)
	q.SetFuzziness(searcher.MaxFuzziness)

	searchResults, err := sm.index.SearchInContext(ctx, bleve.NewSearchRequest(q))
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "performing search query")
	}

	for _, result := range searchResults.Hits {
		ids = append(ids, result.ID)
	}

	return ids, nil
}

// Search implements our IndexManager interface.
func (sm *bleveIndexManager) Search(ctx context.Context, query, accountID string) (ids []string, err error) {
	return sm.search(ctx, query, accountID, false)
}

// SearchForAdmin implements our IndexManager interface.
func (sm *bleveIndexManager) SearchForAdmin(ctx context.Context, query string) (ids []string, err error) {
	return sm.search(ctx, query, "", true)
}

// Delete implements our IndexManager interface.
func (sm *bleveIndexManager) Delete(ctx context.Context, id string) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	logger := sm.logger.WithValue("id", id)

	if err := sm.index.Delete(id); err != nil {
		return observability.PrepareError(err, logger, span, "removing from index")
	}

	sm.logger.WithValue("id", id).Debug("removed from index")

	return nil
}
