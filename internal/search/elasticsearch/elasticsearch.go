package elasticsearch

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic/v7"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
)

var _ search.IndexManager = (*indexManager)(nil)

type (
	indexManager struct {
		indexName    string
		searchFields []string
		esclient     *elastic.Client
		logger       logging.Logger
		tracer       tracing.Tracer
	}
)

// NewIndexManager instantiates an Elasticsearch client.
func NewIndexManager(ctx context.Context, logger logging.Logger, path search.IndexPath, name search.IndexName, fields ...string) (search.IndexManager, error) {
	l := logger.WithName("search")

	client, err := elastic.NewClient(
		elastic.SetURL(string(path)),
		elastic.SetErrorLog(l),
		elastic.SetInfoLog(l),
		elastic.SetTraceLog(l),
	)
	if err != nil {
		return nil, err
	}

	_, _, err = client.Ping(string(path)).Do(ctx)
	if err != nil {
		return nil, err
	}

	im := &indexManager{
		indexName:    string(name),
		esclient:     client,
		logger:       l,
		searchFields: fields,
		tracer:       tracing.NewTracer("search"),
	}

	if err = im.ensureIndices(ctx); err != nil {
		return nil, err
	}

	return im, nil
}

const (
	itemsIndexName = "items"
)

func (sm *indexManager) ensureIndices(ctx context.Context) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	indexExists, err := sm.esclient.IndexExists(itemsIndexName).Do(ctx)
	if err != nil {
		return err
	}

	if !indexExists {
		_, err = sm.esclient.CreateIndex(itemsIndexName).Do(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Index implements our IndexManager interface.
func (sm *indexManager) Index(ctx context.Context, id string, value interface{}) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	sm.logger.WithValue("id", id).Debug("adding to index")

	_, err := sm.esclient.Index().Index(sm.indexName).Id(id).BodyJson(value).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

type idContainer struct {
	ID string `json:"id"`
}

// search executes search queries.
func (sm *indexManager) search(ctx context.Context, query, accountID string, forServiceAdmin bool, fields ...string) (ids []string, err error) {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, query)
	logger := sm.logger.WithValue(keys.SearchQueryKey, query)

	if query == "" {
		return nil, search.ErrEmptyQueryProvided
	}

	baseQuery := elastic.NewMultiMatchQuery(query, fields...).Operator("OR")

	var q elastic.Query
	if accountID == "" && forServiceAdmin {
		q = baseQuery
	} else {
		accountIDMatchQuery := elastic.NewMatchQuery("accountID", accountID)
		q = elastic.NewBoolQuery().Should(accountIDMatchQuery).Should(baseQuery)
	}

	results, err := sm.esclient.Search().Index(sm.indexName).Query(q).Do(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying elasticsearch")
	}

	var returnedItems []string
	for _, hit := range results.Hits.Hits {
		var i *idContainer
		if unmarshalErr := json.Unmarshal(hit.Source, &i); unmarshalErr != nil {
			return nil, observability.PrepareError(err, logger, span, "unmarshaling result")
		}
		returnedItems = append(returnedItems, i.ID)
	}

	return returnedItems, nil
}

// Search implements our IndexManager interface.
func (sm *indexManager) Search(ctx context.Context, query, accountID string) (ids []string, err error) {
	return sm.search(ctx, query, accountID, false)
}

// SearchForAdmin implements our IndexManager interface.
func (sm *indexManager) SearchForAdmin(ctx context.Context, query string) (ids []string, err error) {
	return sm.search(ctx, query, "", true)
}

// Delete implements our IndexManager interface.
func (sm *indexManager) Delete(ctx context.Context, id string) error {
	_, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	logger := sm.logger.WithValue("id", id)

	logger.Debug("removed from index")

	return nil
}
