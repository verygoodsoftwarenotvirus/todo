package httpserver

import (
	"context"
	"fmt"
	"testing"

	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fake"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideItemsServiceSessionInfoFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideItemsServiceSessionInfoFetcher()
	})
}

func TestProvideItemsServiceItemIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideItemsServiceItemIDFetcher(noop.NewLogger())
	})
}

func TestProvideUsersServiceUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideUsersServiceUserIDFetcher(noop.NewLogger())
	})
}

func TestProvideUsersServiceSessionInfoFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideUsersServiceSessionInfoFetcher()
	})
}

func TestProvideWebhooksServiceUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideWebhooksServiceUserIDFetcher()
	})
}

func TestProvideWebhooksServiceWebhookIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideWebhooksServiceWebhookIDFetcher(noop.NewLogger())
	})
}

func TestProvideOAuth2ClientsServiceClientIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideOAuth2ClientsServiceClientIDFetcher(noop.NewLogger())
	})
}

func Test_userIDFetcherFromRequestContext(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakemodels.BuildFakeUser()
		expected := exampleUser.ToSessionInfo()

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, expected),
		)

		actual := userIDFetcherFromRequestContext(req)
		assert.Equal(t, expected.UserID, actual)
	})

	T.Run("without attached value", func(t *testing.T) {
		t.Parallel()

		req := buildRequest(t)
		actual := userIDFetcherFromRequestContext(req)

		assert.Zero(t, actual)
	})
}

func Test_buildRouteParamUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fn := buildRouteParamUserIDFetcher(noop.NewLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{usersservice.UserIDURIParamKey},
						Values: []string{fmt.Sprintf("%d", expected)},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value somehow", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		// NOTE: This will probably never happen in dev or production
		fn := buildRouteParamUserIDFetcher(noop.NewLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{usersservice.UserIDURIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildRouteParamItemIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fn := buildRouteParamItemIDFetcher(noop.NewLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{itemsservice.ItemIDURIParamKey},
						Values: []string{fmt.Sprintf("%d", expected)},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value somehow", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		// NOTE: This will probably never happen in dev or production
		fn := buildRouteParamItemIDFetcher(noop.NewLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{itemsservice.ItemIDURIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildRouteParamWebhookIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fn := buildRouteParamWebhookIDFetcher(noop.NewLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{webhooksservice.WebhookIDURIParamKey},
						Values: []string{fmt.Sprintf("%d", expected)},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value somehow", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		// NOTE: This will probably never happen in dev or production
		fn := buildRouteParamWebhookIDFetcher(noop.NewLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{webhooksservice.WebhookIDURIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildRouteParamOAuth2ClientIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		fn := buildRouteParamOAuth2ClientIDFetcher(noop.NewLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{oauth2clientsservice.OAuth2ClientIDURIParamKey},
						Values: []string{fmt.Sprintf("%d", expected)},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value somehow", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		// NOTE: This will probably never happen in dev or production
		fn := buildRouteParamOAuth2ClientIDFetcher(noop.NewLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{oauth2clientsservice.OAuth2ClientIDURIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}
