package httpclient

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

	const expectedPathFormat = "/api/v1/_admin_/users/status"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		actual, err := c.BuildAccountStatusUpdateInputRequest(ctx, exampleInput)
		assert.NoError(t, err)
		require.NotNil(t, actual)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BanUser(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/_admin_/users/status"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.UpdateAccountStatus(ctx, exampleInput)
		assert.NoError(t, err)
	})

	T.Run("with bad request response", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusBadRequest)
				},
			),
		)
		c := buildTestClient(t, ts)

		assert.Error(t, c.UpdateAccountStatus(ctx, exampleInput))
	})

	T.Run("with otherwise invalid status code response", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.UpdateAccountStatus(ctx, exampleInput)
		assert.Error(t, err)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		c := buildTestClientWithInvalidURL(t)
		err := c.UpdateAccountStatus(ctx, exampleInput)

		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

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
		c.SetOption(UsingTimeout(time.Millisecond))

		err := c.UpdateAccountStatus(ctx, exampleInput)
		assert.Error(t, err)
	})
}
