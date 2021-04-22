package chi

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func Test_sessionContextDataFetcherFromSessionContextData(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := &chiRouteParamManager{}

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		expected, _ := types.SessionContextDataFromUser(exampleUser, exampleAccount.ID, map[uint64]*types.UserAccountMembershipInfo{
			exampleAccount.ID: {
				AccountName: exampleAccount.Name,
				Permissions: testutil.BuildMaxUserPerms(),
			},
		})

		req := testutil.BuildTestRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionContextDataKey, expected),
		)

		actual, err := r.FetchContextFromRequest(req)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("without attached value", func(t *testing.T) {
		t.Parallel()

		r := &chiRouteParamManager{}

		req := testutil.BuildTestRequest(t)
		actual, err := r.FetchContextFromRequest(req)

		assert.Error(t, err)
		assert.Zero(t, actual)
	})
}

func Test_BuildRouteParamIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := &chiRouteParamManager{}

		ctx := context.Background()
		exampleKey := "blah"
		fn := r.BuildRouteParamIDFetcher(logging.NewNonOperationalLogger(), exampleKey, "thing")
		expected := uint64(123)
		req := testutil.BuildTestRequest(t).WithContext(
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

		r := &chiRouteParamManager{}

		ctx := context.Background()
		exampleKey := "blah"
		fn := r.BuildRouteParamIDFetcher(logging.NewNonOperationalLogger(), exampleKey, "thing")
		expected := uint64(0)

		req := testutil.BuildTestRequest(t)
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
