package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildGetWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := h.builder.BuildGetWebhookRequest(h.ctx, exampleWebhook.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetWebhookRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetWebhooksRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := h.builder.BuildGetWebhooksRequest(h.ctx, nil)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeWebhookCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := h.builder.BuildCreateWebhookRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateWebhookRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		h.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := &types.WebhookCreationInput{}

		actual, err := h.builder.BuildCreateWebhookRequest(h.ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildUpdateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := h.builder.BuildUpdateWebhookRequest(h.ctx, exampleWebhook)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUpdateWebhookRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodDelete, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := h.builder.BuildArchiveWebhookRequest(h.ctx, exampleWebhook.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildArchiveWebhookRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := h.builder.BuildGetAuditLogForWebhookRequest(h.ctx, exampleWebhook.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleWebhook.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForWebhookRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
