package httpserver

import (
	"context"
	"fmt"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func TestProvideItemServiceUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideItemServiceUserIDFetcher()
	})
}

func TestProvideItemIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideItemIDFetcher(noop.ProvideNoopLogger())
	})
}

func TestProvideUsernameFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideUsernameFetcher(noop.ProvideNoopLogger())
	})
}

func TestProvideAuthUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideAuthUserIDFetcher()
	})
}

func TestProvideWebhooksUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideWebhooksUserIDFetcher()
	})
}

func TestProvideWebhookIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideWebhookIDFetcher(noop.ProvideNoopLogger())
	})
}

func TestProvideOAuth2ServiceClientIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = ProvideOAuth2ServiceClientIDFetcher(noop.ProvideNoopLogger())
	})
}

func TestUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				expected,
			),
		)

		actual := UserIDFetcher(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("without attached value", func(t *testing.T) {
		req := buildRequest(t)
		actual := UserIDFetcher(req)

		assert.Zero(t, actual)
	})
}

func Test_buildChiUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		fn := buildChiUserIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{usersservice.URIParamKey},
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
		fn := buildChiUserIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{usersservice.URIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildChiItemIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		fn := buildChiItemIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{itemsservice.URIParamKey},
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
		fn := buildChiItemIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{itemsservice.URIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildChiWebhookIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		fn := buildChiWebhookIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{webhooksservice.URIParamKey},
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
		fn := buildChiWebhookIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{webhooksservice.URIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildChiOAuth2ClientIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		fn := buildChiOAuth2ClientIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(123)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{oauth2clientsservice.URIParamKey},
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
		fn := buildChiOAuth2ClientIDFetcher(noop.ProvideNoopLogger())
		expected := uint64(0)

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{oauth2clientsservice.URIParamKey},
						Values: []string{"expected"},
					},
				},
			),
		)

		actual := fn(req)
		assert.Equal(t, expected, actual)
	})
}
