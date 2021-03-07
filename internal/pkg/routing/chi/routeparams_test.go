package chi

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)

	return req
}

func Test_userIDFetcherFromRequestContext(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		r := &chirouteParamManager{}

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		expected, _ := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, expected),
		)

		actual := r.UserIDFetcherFromRequestContext(req)
		assert.Equal(t, expected.User.ID, actual)
	})

	T.Run("without attached value", func(t *testing.T) {
		t.Parallel()

		r := &chirouteParamManager{}

		req := buildRequest(t)
		actual := r.UserIDFetcherFromRequestContext(req)

		assert.Zero(t, actual)
	})
}

func Test_requestContextFetcherFromRequestContext(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		r := &chirouteParamManager{}

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		expected, _ := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, expected),
		)

		actual, err := r.FetchContextFromRequest(req)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("without attached value", func(t *testing.T) {
		t.Parallel()

		r := &chirouteParamManager{}

		req := buildRequest(t)
		actual, err := r.FetchContextFromRequest(req)

		assert.Error(t, err)
		assert.Zero(t, actual)
	})
}

func Test_BuildRouteParamIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		r := &chirouteParamManager{}

		ctx := context.Background()
		exampleKey := "blah"
		fn := r.BuildRouteParamIDFetcher(logging.NewNonOperationalLogger(), exampleKey, "thing")
		expected := uint64(123)
		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{exampleKey},
						Values: []string{fmt.Sprintf("%d", expected)},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value somehow", func(t *testing.T) {
		// NOTE: This will probably never happen in dev or production
		t.Parallel()

		r := &chirouteParamManager{}

		ctx := context.Background()
		exampleKey := "blah"
		fn := r.BuildRouteParamIDFetcher(logging.NewNonOperationalLogger(), exampleKey, "thing")
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{exampleKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}
