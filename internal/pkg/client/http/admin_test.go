package client

import (
	"context"
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

	const expectedPathFormat = "/api/v1/_admin_/users/%d/ban"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUserID := fakes.BuildFakeUser().ID
		spec := newRequestSpec(true, http.MethodDelete, expectedPathFormat, exampleUserID)

		actual, err := c.BuildBanUserRequest(ctx, exampleUserID)
		assert.NoError(t, err)
		require.NotNil(t, actual)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BanUser(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/_admin_/users/%d/ban"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeUser().ID
		spec := newRequestSpec(true, http.MethodDelete, expectedPathFormat, exampleUserID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.BanUser(ctx, exampleUserID)
		assert.NoError(t, err)
	})

	T.Run("with bad request response", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeUser().ID
		spec := newRequestSpec(true, http.MethodDelete, expectedPathFormat, exampleUserID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusBadRequest)
				},
			),
		)
		c := buildTestClient(t, ts)

		assert.Error(t, c.BanUser(ctx, exampleUserID))
	})

	T.Run("with otherwise invalid status code response", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeUser().ID
		spec := newRequestSpec(true, http.MethodDelete, expectedPathFormat, exampleUserID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.BanUser(ctx, exampleUserID)
		assert.Error(t, err)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeUser().ID
		c := buildTestClientWithInvalidURL(t)
		err := c.BanUser(ctx, exampleUserID)

		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeUser().ID
		spec := newRequestSpec(true, http.MethodDelete, expectedPathFormat, exampleUserID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					time.Sleep(10 * time.Minute)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = time.Millisecond

		err := c.BanUser(ctx, exampleUserID)
		assert.Error(t, err)
	})
}
