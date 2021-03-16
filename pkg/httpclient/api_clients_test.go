package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_GetAPIClient(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAPIClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAPIClient))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAPIClient(ctx, exampleAPIClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAPIClient, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAPIClient(ctx, exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAPIClients(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		for i := 0; i < len(exampleAPIClientList.Clients); i++ {
			exampleAPIClientList.Clients[i].ClientSecret = nil
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAPIClientList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAPIClients(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAPIClientList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAPIClients(ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_CreateAPIClient(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleResponse := fakes.BuildFakeAPIClientCreationResponseFromClient(exampleAPIClient)
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		actual, err := c.CreateAPIClient(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateAPIClient(ctx, &http.Cookie{}, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response from server", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					_, err := res.Write([]byte("BLAH"))
					assert.NoError(t, err)
				},
			),
		)
		c := buildTestClient(t, ts)

		actual, err := c.CreateAPIClient(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		_, err := c.CreateAPIClient(ctx, nil, nil)
		assert.Error(t, err)
	})
}

func TestV1Client_ArchiveAPIClient(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAPIClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveAPIClient(ctx, exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		err := buildTestClientWithInvalidURL(t).ArchiveAPIClient(ctx, exampleAPIClient.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAuditLogForAPIClient(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/api_clients/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleAPIClient.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAPIClient(ctx, exampleAPIClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAPIClient(ctx, exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleAPIClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAPIClient(ctx, exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
