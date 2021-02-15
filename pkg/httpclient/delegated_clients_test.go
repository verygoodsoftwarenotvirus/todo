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

func TestV1Client_BuildGetDelegatedClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/delegated_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleDelegatedClient.ID)

		actual, err := c.BuildGetDelegatedClientRequest(ctx, exampleDelegatedClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetDelegatedClient(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/delegated_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.HMACKey = nil
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleDelegatedClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleDelegatedClient))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetDelegatedClient(ctx, exampleDelegatedClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleDelegatedClient, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.HMACKey = nil

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetDelegatedClient(ctx, exampleDelegatedClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetDelegatedClientsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/delegated_clients"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := c.BuildGetDelegatedClientsRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetDelegatedClients(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/delegated_clients"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClientList := fakes.BuildFakeDelegatedClientList()
		for i := 0; i < len(exampleDelegatedClientList.Clients); i++ {
			exampleDelegatedClientList.Clients[i].HMACKey = nil
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleDelegatedClientList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetDelegatedClients(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleDelegatedClientList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetDelegatedClients(ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreateDelegatedClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/delegated_client"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		req, err := c.BuildCreateDelegatedClientRequest(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, req, spec)
	})
}

func TestV1Client_CreateDelegatedClient(T *testing.T) {
	T.Parallel()

	const expectedPath = "/delegated_client"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.HMACKey = nil
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleDelegatedClient))
				},
			),
		)
		c := buildTestClient(t, ts)

		actual, err := c.CreateDelegatedClient(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleDelegatedClient, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateDelegatedClient(ctx, &http.Cookie{}, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response from server", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

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

		actual, err := c.CreateDelegatedClient(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		_, err := c.CreateDelegatedClient(ctx, nil, nil)
		assert.Error(t, err)
	})
}

func TestV1Client_BuildArchiveDelegatedClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/delegated_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleDelegatedClient.ID)
		c := buildTestClient(t, ts)

		actual, err := c.BuildArchiveDelegatedClientRequest(ctx, exampleDelegatedClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_ArchiveDelegatedClient(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/delegated_clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleDelegatedClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveDelegatedClient(ctx, exampleDelegatedClient.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		err := buildTestClientWithInvalidURL(t).ArchiveDelegatedClient(ctx, exampleDelegatedClient.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetAuditLogForDelegatedClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/delegated_clients/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForDelegatedClientRequest(ctx, exampleDelegatedClient.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleDelegatedClient.ID)
		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAuditLogForDelegatedClient(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/delegated_clients/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleDelegatedClient.ID)
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
		actual, err := c.GetAuditLogForDelegatedClient(ctx, exampleDelegatedClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForDelegatedClient(ctx, exampleDelegatedClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleDelegatedClient.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForDelegatedClient(ctx, exampleDelegatedClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
