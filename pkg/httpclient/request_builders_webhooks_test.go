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

func TestV1Client_BuildGetWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, exampleWebhook.ID)

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetWebhookRequest(ctx, exampleWebhook.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetWebhooksRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetWebhooksRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildCreateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		c := buildTestClient(t, ts)

		actual, err := c.BuildCreateWebhookRequest(ctx, exampleInput)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildUpdateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleWebhook.ID)

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildUpdateWebhookRequest(ctx, exampleWebhook)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildArchiveWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleWebhook := fakes.BuildFakeWebhook()
		spec := newRequestSpec(false, http.MethodDelete, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := c.BuildArchiveWebhookRequest(ctx, exampleWebhook.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAuditLogForWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleWebhook := fakes.BuildFakeWebhook()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForWebhookRequest(ctx, exampleWebhook.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleWebhook.ID)
		assertRequestQuality(t, actual, spec)
	})
}
