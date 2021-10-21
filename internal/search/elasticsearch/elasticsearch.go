package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
)

var _ search.IndexManager = (*indexManager)(nil)

type (
	esClient interface {
		IndexExists(indices ...string) *elastic.IndicesExistsService
		CreateIndex(name string) *elastic.IndicesCreateService
		Search(indices ...string) *elastic.SearchService
		Index() *elastic.IndexService
		DeleteByQuery(indices ...string) *elastic.DeleteByQueryService
	}

	indexManager struct {
		logger       logging.Logger
		tracer       tracing.Tracer
		esclient     esClient
		indexName    string
		searchFields []string
	}
)

// NewIndexManager instantiates an Elasticsearch client.
func NewIndexManager(
	ctx context.Context,
	logger logging.Logger,
	client *http.Client,
	path search.IndexPath,
	name search.IndexName,
	fields ...string,
) (search.IndexManager, error) {
	l := logger.WithName("search")

	if !elasticsearchIsReady(client, path, logger, 50) {
		return nil, errors.New("elasticsearch isn't ready")
	}

	c, err := elastic.NewClient(
		elastic.SetURL(string(path)),
		elastic.SetHttpClient(client),
		elastic.SetHealthcheck(false),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	im := &indexManager{
		indexName:    string(name),
		esclient:     c,
		logger:       l,
		searchFields: fields,
		tracer:       tracing.NewTracer("search"),
	}

	if indexErr := im.ensureIndices(ctx, name); indexErr != nil {
		return nil, indexErr
	}

	return im, nil
}

func elasticsearchIsReady(
	client *http.Client,
	path search.IndexPath,
	l logging.Logger,
	maxAttempts uint8,
) (ready bool) {
	attemptCount := 0

	logger := l.WithValues(map[string]interface{}{
		"interval":     time.Second.String(),
		"max_attempts": maxAttempts,
	})

	for !ready {
		_, err := elastic.NewClient(
			elastic.SetURL(string(path)),
			elastic.SetHttpClient(client),
		)
		if err != nil {
			logger.WithValue("attempt_count", attemptCount).Debug("ping failed, waiting for elasticsearch")
			time.Sleep(time.Second)

			attemptCount++
			if attemptCount >= int(maxAttempts) {
				break
			}
		} else {
			ready = true
			return ready
		}
	}

	return false
}

func (sm *indexManager) ensureIndices(ctx context.Context, indices ...search.IndexName) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	logger := sm.logger.WithValue("indices", indices)

	for _, index := range indices {
		indexExists, err := sm.esclient.IndexExists(strings.ToLower(string(index))).Do(ctx)
		if err != nil {
			return observability.PrepareError(err, logger, span, "checking index existence")
		}

		if !indexExists {
			if _, err = sm.esclient.CreateIndex(strings.ToLower(string(index))).Do(ctx); err != nil && !strings.Contains(elastic.ErrorReason(err), "already exists") {
				return observability.PrepareError(err, logger, span, elastic.ErrorReason(err))
			}
		}
	}

	return nil
}

// Index implements our IndexManager interface.
func (sm *indexManager) Index(ctx context.Context, id string, value interface{}) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	logger := sm.logger.WithValue("id", id).WithValue("value", value)
	logger.Debug("adding to index")

	if _, err := sm.esclient.Index().Index(sm.indexName).Id(id).BodyJson(value).Do(ctx); err != nil {
		return observability.PrepareError(err, logger, span, "indexing value")
	}

	return nil
}

type idContainer struct {
	ID string `json:"id"`
}

var (
	// ErrEmptyQueryProvided indicates an empty query was provided as input.
	ErrEmptyQueryProvided = errors.New("empty search query provided")
)

// search executes search queries.
func (sm *indexManager) search(
	ctx context.Context,
	query,
	accountID string,
) (ids []string, err error) {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, query)
	logger := sm.logger.WithValue(keys.SearchQueryKey, query)

	if query == "" {
		return nil, ErrEmptyQueryProvided
	}

	baseQuery := elastic.NewMultiMatchQuery(query, sm.searchFields...)

	var q elastic.Query
	if accountID == "" {
		q = baseQuery
	} else {
		accountIDMatchQuery := elastic.NewMatchQuery("accountID", accountID)
		q = elastic.NewBoolQuery().Should(accountIDMatchQuery).Should(baseQuery)
	}

	results, err := sm.esclient.Search().Index(sm.indexName).Query(q).Do(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying elasticsearch")
	}

	resultIDs := []string{}
	for _, hit := range results.Hits.Hits {
		var i *idContainer
		if unmarshalErr := json.Unmarshal(hit.Source, &i); unmarshalErr != nil {
			return nil, observability.PrepareError(err, logger, span, "unmarshalling search result")
		}
		resultIDs = append(resultIDs, i.ID)
	}

	return resultIDs, nil
}

// Search implements our IndexManager interface.
func (sm *indexManager) Search(ctx context.Context, query, accountID string) (ids []string, err error) {
	return sm.search(ctx, query, accountID)
}

// Delete implements our IndexManager interface.
func (sm *indexManager) Delete(ctx context.Context, id string) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	logger := sm.logger.WithValue("id", id)

	q := elastic.NewTermQuery("id", id)
	if _, err := sm.esclient.DeleteByQuery(sm.indexName).Query(q).Do(ctx); err != nil {
		return observability.PrepareError(err, logger, span, "deleting from elasticsearch")
	}

	logger.Debug("removed from index")

	return nil
}
