package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildItemExistsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleItem := fakes.BuildFakeItem()

		actual, err := helper.builder.BuildItemExistsRequest(helper.ctx, exampleItem.ID)
		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, exampleItem.ID)

		assert.NoError(t, err)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildItemExistsRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleItem := fakes.BuildFakeItem()

		actual, err := helper.builder.BuildItemExistsRequest(helper.ctx, exampleItem.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleItem.ID)

		actual, err := helper.builder.BuildGetItemRequest(helper.ctx, exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetItemRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleItem := fakes.BuildFakeItem()

		actual, err := helper.builder.BuildGetItemRequest(helper.ctx, exampleItem.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPathFormat)

		actual, err := helper.builder.BuildGetItemsRequest(helper.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		filter := (*types.QueryFilter)(nil)

		actual, err := helper.builder.BuildGetItemsRequest(helper.ctx, filter)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildSearchItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/search"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		limit := types.DefaultQueryFilter().Limit
		exampleQuery := "whatever"
		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

		actual, err := helper.builder.BuildSearchItemsRequest(helper.ctx, exampleQuery, limit)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		limit := types.DefaultQueryFilter().Limit
		exampleQuery := "whatever"

		actual, err := helper.builder.BuildSearchItemsRequest(helper.ctx, exampleQuery, limit)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildCreateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeItemCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCreateItemRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateItemRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateItemRequest(helper.ctx, &types.ItemCreationInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleInput := fakes.BuildFakeItemCreationInput()

		actual, err := helper.builder.BuildCreateItemRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildUpdateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleItem.ID)

		actual, err := helper.builder.BuildUpdateItemRequest(helper.ctx, exampleItem)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildUpdateItemRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleItem := fakes.BuildFakeItem()

		actual, err := helper.builder.BuildUpdateItemRequest(helper.ctx, exampleItem)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleItem := fakes.BuildFakeItem()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleItem.ID)

		actual, err := helper.builder.BuildArchiveItemRequest(helper.ctx, exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveItemRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleItem := fakes.BuildFakeItem()

		actual, err := helper.builder.BuildArchiveItemRequest(helper.ctx, exampleItem.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
