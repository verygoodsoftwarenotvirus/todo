package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildGetAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAPIClient.ID)

		actual, err := helper.builder.BuildGetAPIClientRequest(helper.ctx, exampleAPIClient.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAPIClientRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAPIClientsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetAPIClientsRequest(helper.ctx, nil)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeAPIClientCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCreateAPIClientRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeAPIClientCreationInput()

		actual, err := helper.builder.BuildCreateAPIClientRequest(helper.ctx, nil, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAPIClientRequest(helper.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building data request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeAPIClientCreationInput()

		helper.builder = buildTestRequestBuilderWithInvalidURL()
		actual, err := helper.builder.BuildCreateAPIClientRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/api_clients/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAPIClient.ID)

		actual, err := helper.builder.BuildArchiveAPIClientRequest(helper.ctx, exampleAPIClient.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveAPIClientRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForAPIClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/api_clients/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		actual, err := helper.builder.BuildGetAuditLogForAPIClientRequest(helper.ctx, exampleAPIClient.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAPIClient.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForAPIClientRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
