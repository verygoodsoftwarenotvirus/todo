package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetWebhookRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		expectedID := uint64(1)

		actual, err := c.BuildGetWebhookRequest(ctx, expectedID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:   1,
			Name: "example",
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t,
						strings.HasSuffix(
							req.URL.String(),
							strconv.Itoa(int(expected.ID)),
						),
					)
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/webhooks/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetWebhook(ctx, expected.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildGetWebhooksRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetWebhooksRequest(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.WebhookList{
			Webhooks: []models.Webhook{
				{
					ID:   1,
					Name: "example",
				},
			},
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/webhooks", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetWebhooks(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildCreateWebhookRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodPost
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleInput := &models.WebhookCreationInput{
			Name: "expected name",
		}
		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateWebhookRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:   1,
			Name: "example",
		}

		exampleInput := &models.WebhookCreationInput{
			Name: expected.Name,
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/webhooks", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					var x *models.WebhookCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(expected))
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.CreateWebhook(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildUpdateWebhookRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodPut
		ctx := context.Background()

		exampleInput := &models.Webhook{
			Name: "changed name",
		}

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateWebhookRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:   1,
			Name: "example",
		}
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/webhooks/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPut)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).UpdateWebhook(ctx, expected)

		assert.NoError(t, err, "no error should be returned")
	})
}

func TestV1Client_BuildArchiveWebhookRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodDelete
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		expectedID := uint64(1)
		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveWebhookRequest(ctx, expectedID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(1)
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/webhooks/%d", expected), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveWebhook(ctx, expected)

		assert.NoError(t, err, "no error should be returned")
	})
}
