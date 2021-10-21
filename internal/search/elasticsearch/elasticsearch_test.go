package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

type mockESClient struct {
	mock.Mock
}

func (m *mockESClient) IndexExists(indices ...string) *elastic.IndicesExistsService {
	return m.Called(indices).Get(0).(*elastic.IndicesExistsService)
}

func (m *mockESClient) CreateIndex(name string) *elastic.IndicesCreateService {
	return m.Called(name).Get(0).(*elastic.IndicesCreateService)
}

func (m *mockESClient) Search(indices ...string) *elastic.SearchService {
	return m.Called(indices).Get(0).(*elastic.SearchService)
}

func (m *mockESClient) Index() *elastic.IndexService {
	return m.Called().Get(0).(*elastic.IndexService)
}

func (m *mockESClient) DeleteByQuery(indices ...string) *elastic.DeleteByQueryService {
	return m.Called(indices).Get(0).(*elastic.DeleteByQueryService)
}

func TestNewIndexManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		//
	})
}

const exampleIndexName = "example"

func Test_indexManager_ensureIndices(T *testing.T) {
	T.Parallel()

	T.Run("standard with existent index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indicesExistsService := elastic.NewIndicesExistsService(client).Index([]string{exampleIndexName})

		esc := &mockESClient{}
		esc.On("IndexExists", []string{exampleIndexName}).Return(indicesExistsService)

		im := &indexManager{
			esclient: esc,
			tracer:   tracing.NewTracer(t.Name()),
			logger:   logger,
		}
		assert.NoError(t, im.ensureIndices(ctx, exampleIndexName))

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("standard with nonexistent index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					res.WriteHeader(http.StatusNotFound)
					return
				}

				_, err := res.Write([]byte(`{}`))
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indicesExistsService := elastic.NewIndicesExistsService(client).Index([]string{exampleIndexName})
		indicesCreateService := elastic.NewIndicesCreateService(client).Index(exampleIndexName)

		esc := &mockESClient{}
		esc.On("IndexExists", []string{exampleIndexName}).Return(indicesExistsService)
		esc.On("CreateIndex", exampleIndexName).Return(indicesCreateService)

		im := &indexManager{
			esclient: esc,
			tracer:   tracing.NewTracer(t.Name()),
			logger:   logger,
		}
		assert.NoError(t, im.ensureIndices(ctx, exampleIndexName))

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("with error checking index existence", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusInternalServerError)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indicesExistsService := elastic.NewIndicesExistsService(client).Index([]string{exampleIndexName})

		esc := &mockESClient{}
		esc.On("IndexExists", []string{exampleIndexName}).Return(indicesExistsService)

		im := &indexManager{
			esclient: esc,
			tracer:   tracing.NewTracer(t.Name()),
			logger:   logger,
		}
		assert.Error(t, im.ensureIndices(ctx, exampleIndexName))

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("with error creating nonexistent index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					res.WriteHeader(http.StatusNotFound)
					return
				}

				res.WriteHeader(http.StatusInternalServerError)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indicesExistsService := elastic.NewIndicesExistsService(client).Index([]string{exampleIndexName})
		indicesCreateService := elastic.NewIndicesCreateService(client).Index(exampleIndexName)

		esc := &mockESClient{}
		esc.On("IndexExists", []string{exampleIndexName}).Return(indicesExistsService)
		esc.On("CreateIndex", exampleIndexName).Return(indicesCreateService)

		im := &indexManager{
			esclient: esc,
			tracer:   tracing.NewTracer(t.Name()),
			logger:   logger,
		}
		assert.Error(t, im.ensureIndices(ctx, exampleIndexName))

		mock.AssertExpectationsForObjects(t, esc)
	})
}

func Test_indexManager_Index(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				_, err := res.Write([]byte("{}"))
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indicesExistsService := elastic.NewIndexService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Index").Return(indicesExistsService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}
		assert.NoError(t, im.Index(ctx, t.Name(), t.Name()))

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("with error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusInternalServerError)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		indexService := elastic.NewIndexService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Index").Return(indexService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}
		assert.Error(t, im.Index(ctx, t.Name(), t.Name()))

		mock.AssertExpectationsForObjects(t, esc)
	})
}

func Test_indexManager_search(T *testing.T) {
	T.Parallel()

	T.Run("standard for non-admin", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				results := &elastic.SearchResult{
					Hits: &elastic.SearchHits{
						Hits: []*elastic.SearchHit{
							{
								Source: []byte(fmt.Sprintf(`{"id": %q}`, t.Name())),
							},
						},
					},
				}
				output, err := json.Marshal(results)
				require.NoError(t, err)

				_, err = res.Write(output)
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		searchService := elastic.NewSearchService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Search", []string(nil)).Return(searchService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}

		results, err := im.search(ctx, t.Name(), t.Name())
		assert.NotNil(t, results)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("with invalid query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		im := &indexManager{
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}

		results, err := im.search(ctx, "", t.Name())
		assert.Nil(t, results)
		assert.Error(t, err)
	})

	T.Run("with error executing search", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusInternalServerError)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		searchService := elastic.NewSearchService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Search", []string(nil)).Return(searchService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}

		results, err := im.search(ctx, t.Name(), t.Name())
		assert.Nil(t, results)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, esc)
	})

	// NOTE: This would ideally be a test that validates when the service returns invalid JSON that
	// decoding fails, only that doesn't yield the error in the expected place, so no additional code is covered.
	T.Run("with invalid results", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		replaceString := `{"REPLACEME":true}`

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				results := &elastic.SearchResult{
					Hits: &elastic.SearchHits{
						Hits: []*elastic.SearchHit{
							{
								Source: []byte(replaceString),
							},
						},
					},
				}
				preoutput, err := json.Marshal(results)
				require.NoError(t, err)

				output := []byte(strings.ReplaceAll(string(preoutput), replaceString, `} invalid JSON lmao`))

				_, err = res.Write(output)
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		searchService := elastic.NewSearchService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Search", []string(nil)).Return(searchService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}

		results, err := im.search(ctx, t.Name(), t.Name())
		assert.Nil(t, results)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, esc)
	})
}

func Test_indexManager_Search(T *testing.T) {
	T.Parallel()

	T.Run("standard for non-admin", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				results := &elastic.SearchResult{
					Hits: &elastic.SearchHits{
						Hits: []*elastic.SearchHit{
							{
								Source: []byte(fmt.Sprintf(`{"id": %q}`, t.Name())),
							},
						},
					},
				}
				output, err := json.Marshal(results)
				require.NoError(t, err)

				_, err = res.Write(output)
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		searchService := elastic.NewSearchService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("Search", []string(nil)).Return(searchService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}

		results, err := im.Search(ctx, t.Name(), t.Name())
		assert.NotNil(t, results)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, esc)
	})
}

func Test_indexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				_, err := res.Write([]byte("{}"))
				require.NoError(t, err)

				res.WriteHeader(http.StatusOK)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		deletionService := elastic.NewDeleteByQueryService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("DeleteByQuery", []string{t.Name()}).Return(deletionService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}
		assert.NoError(t, im.Delete(ctx, t.Name()))

		mock.AssertExpectationsForObjects(t, esc)
	})

	T.Run("with error deleting", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewZerologLogger()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusInternalServerError)
			},
		))

		client, err := elastic.NewSimpleClient(
			elastic.SetHttpClient(ts.Client()),
			elastic.SetURL(ts.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)

		deletionService := elastic.NewDeleteByQueryService(client).Index(t.Name())

		esc := &mockESClient{}
		esc.On("DeleteByQuery", []string{t.Name()}).Return(deletionService)

		im := &indexManager{
			esclient:  esc,
			indexName: t.Name(),
			tracer:    tracing.NewTracer(t.Name()),
			logger:    logger,
		}
		assert.Error(t, im.Delete(ctx, t.Name()))

		mock.AssertExpectationsForObjects(t, esc)
	})
}
