package client

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

func TestV1Client_BuildGetAuditLogEntriesRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/_admin_/audit_log"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogEntriesRequest(ctx, filter)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/_admin_/audit_log"
		expectedMethod = http.MethodGet
	)

	spec := newRequestSpec(true, expectedMethod, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		filter := (*types.QueryFilter)(nil)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogEntries(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogEntries(ctx, filter)

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
		actual, err := c.GetAuditLogEntries(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetAuditLogEntryRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/_admin_/audit_log/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAuditLogEntry.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAuditLogEntryRequest(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/_admin_/audit_log/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAuditLogEntry.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntry))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntry, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAuditLogEntry.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
