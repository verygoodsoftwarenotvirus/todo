package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_GetAccount(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccount.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAccount))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccount(ctx, exampleAccount.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAccount, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccount(ctx, exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccount.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccount(ctx, exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAccounts(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		exampleAccountList := fakes.BuildFakeAccountList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAccountList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccounts(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAccountList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccounts(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccounts(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_CreateAccount(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = 0
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					var x *types.AccountCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAccount))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.CreateAccount(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAccount, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateAccount(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_UpdateAccount(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccount.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					assert.NoError(t, json.NewEncoder(res).Encode(exampleAccount))
				},
			),
		)

		err := buildTestClient(t, ts).UpdateAccount(ctx, exampleAccount)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()

		err := buildTestClientWithInvalidURL(t).UpdateAccount(ctx, exampleAccount)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_ArchiveAccount(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccount.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveAccount(ctx, exampleAccount.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()

		err := buildTestClientWithInvalidURL(t).ArchiveAccount(ctx, exampleAccount.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAuditLogForAccount(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/accounts/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleAccount.ID)
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
		actual, err := c.GetAuditLogForAccount(ctx, exampleAccount.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAccount(ctx, exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleAccount.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAccount(ctx, exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
