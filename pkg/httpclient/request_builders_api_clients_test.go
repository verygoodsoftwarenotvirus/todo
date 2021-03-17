package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		exampleAPIClient := fakes.BuildFakeAPIClient()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAPIClient.ID)

		actual, err := c.BuildGetAPIClientRequest(ctx, exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAPIClientsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := c.BuildGetAPIClientsRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildCreateAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		req, err := c.requestBuilder.BuildCreateAPIClientRequest(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, req, spec)
	})
}

func TestV1Client_BuildArchiveAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAPIClient.ID)
		c := buildTestClient(t, ts)

		actual, err := c.BuildArchiveAPIClientRequest(ctx, exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAuditLogForAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForAPIClientRequest(ctx, exampleAPIClient.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAPIClient.ID)
		assertRequestQuality(t, actual, spec)
	})
}
