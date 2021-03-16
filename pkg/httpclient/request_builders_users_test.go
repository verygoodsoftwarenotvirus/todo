package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleUser.ID)

		actual, err := c.BuildGetUserRequest(ctx, exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetUsersRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetUsersRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildSearchForUsersByUsernameRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/search"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		spec := newRequestSpec(false, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildSearchForUsersByUsernameRequest(ctx, exampleUsername)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildCreateUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
		c := buildTestClient(t, ts)

		actual, err := c.BuildCreateUserRequest(ctx, exampleInput)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildArchiveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleUser.ID)

		actual, err := c.BuildArchiveUserRequest(ctx, exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAuditLogForUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForUserRequest(ctx, exampleUser.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleUser.ID)
		assertRequestQuality(t, actual, spec)
	})
}
