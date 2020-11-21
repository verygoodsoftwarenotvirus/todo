package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildBanUserRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()

		req, err := c.BuildBanUserRequest(ctx, exampleUser.ID)
		assert.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, req.Method, http.MethodDelete)
	})
}

func TestV1Client_BanUser(T *testing.T) {
	T.Parallel()

	expectedPathFormat := "/api/v1/_admin_/users/%d/ban"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		expectedPath := fmt.Sprintf(expectedPathFormat, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.BanUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
	})

	T.Run("with bad request response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		expectedPath := fmt.Sprintf(expectedPathFormat, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusBadRequest)
				},
			),
		)
		c := buildTestClient(t, ts)

		assert.Error(t, c.BanUser(ctx, exampleUser.ID))
	})

	T.Run("with otherwise invalid status code response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		expectedPath := fmt.Sprintf(expectedPathFormat, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.BanUser(ctx, exampleUser.ID)
		assert.Error(t, err)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		c := buildTestClientWithInvalidURL(t)
		err := c.BanUser(ctx, exampleUser.ID)

		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		expectedPath := fmt.Sprintf(expectedPathFormat, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodDelete)

					time.Sleep(10 * time.Minute)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = time.Millisecond

		err := c.BanUser(ctx, exampleUser.ID)
		assert.Error(t, err)
	})
}
