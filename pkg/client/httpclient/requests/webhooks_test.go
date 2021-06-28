package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildGetWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := helper.builder.BuildGetWebhookRequest(helper.ctx, exampleWebhook.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetWebhookRequest(helper.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := helper.builder.BuildGetWebhookRequest(helper.ctx, exampleWebhook.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetWebhooksRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetWebhooksRequest(helper.ctx, nil)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		actual, err := helper.builder.BuildGetWebhooksRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildCreateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeWebhookCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCreateWebhookRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateWebhookRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := &types.WebhookCreationInput{}

		actual, err := helper.builder.BuildCreateWebhookRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakeWebhookCreationInput()

		actual, err := helper.builder.BuildCreateWebhookRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildUpdateWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := helper.builder.BuildUpdateWebhookRequest(helper.ctx, exampleWebhook)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildUpdateWebhookRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := helper.builder.BuildUpdateWebhookRequest(helper.ctx, exampleWebhook)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/webhooks/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		spec := newRequestSpec(false, http.MethodDelete, "", expectedPathFormat, exampleWebhook.ID)

		actual, err := helper.builder.BuildArchiveWebhookRequest(helper.ctx, exampleWebhook.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveWebhookRequest(helper.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := helper.builder.BuildArchiveWebhookRequest(helper.ctx, exampleWebhook.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetAuditLogForWebhookRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/webhooks/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := helper.builder.BuildGetAuditLogForWebhookRequest(helper.ctx, exampleWebhook.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleWebhook.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForWebhookRequest(helper.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleWebhook := fakes.BuildFakeWebhook()

		actual, err := helper.builder.BuildGetAuditLogForWebhookRequest(helper.ctx, exampleWebhook.ID)
		require.Nil(t, actual)
		assert.Error(t, err)
	})
}
