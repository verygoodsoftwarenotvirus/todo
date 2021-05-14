package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildItemExistsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleItem := fakes.BuildFakeItem()

		actual, err := h.builder.BuildItemExistsRequest(h.ctx, exampleItem.ID)
		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, exampleItem.ID)

		assert.NoError(t, err)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildItemExistsRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleItem.ID)

		actual, err := h.builder.BuildGetItemRequest(h.ctx, exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetItemRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := h.builder.BuildGetItemsRequest(h.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildSearchItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/search"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		limit := types.DefaultQueryFilter().Limit
		exampleQuery := "whatever"
		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

		actual, err := h.builder.BuildSearchItemsRequest(h.ctx, exampleQuery, limit)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeItemCreationInput()

		actual, err := h.builder.BuildCreateItemRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateItemRequest(h.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateItemRequest(h.ctx, &types.ItemCreationInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildUpdateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleItem.ID)

		actual, err := h.builder.BuildUpdateItemRequest(h.ctx, exampleItem)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUpdateItemRequest(h.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleItem.ID)

		actual, err := h.builder.BuildArchiveItemRequest(h.ctx, exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildArchiveItemRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetAuditLogForItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleItem := fakes.BuildFakeItem()

		actual, err := h.builder.BuildGetAuditLogForItemRequest(h.ctx, exampleItem.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleItem.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForItemRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
