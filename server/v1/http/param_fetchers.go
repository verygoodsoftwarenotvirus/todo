package httpserver

import (
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/google/wire"
)

var (
	paramFetcherProviders = wire.NewSet(
		ProvideUserIDFetcher,
		ProvideUsernameFetcher,
		ProvideOAuth2ServiceClientIDFetcher,
		ProvideAuthUserIDFetcher,
		ProvideItemIDFetcher,
		ProvideWebhooksUserIDFetcher,
		ProvideWebhookIDFetcher,
	)
)

// ProvideUserIDFetcher provides a UserIDFetcher
func ProvideUserIDFetcher() items.UserIDFetcher {
	return UserIDFetcher
}

// ProvideItemIDFetcher provides an ItemIDFetcher
func ProvideItemIDFetcher() items.ItemIDFetcher {
	return chiItemIDFetcher
}

// ProvideUsernameFetcher provides a UsernameFetcher
func ProvideUsernameFetcher() users.UserIDFetcher {
	return ChiUserIDFetcher
}

// ProvideAuthUserIDFetcher provides a UsernameFetcher
func ProvideAuthUserIDFetcher() auth.UserIDFetcher {
	return UserIDFetcher
}

// ProvideWebhooksUserIDFetcher provides a UserIDFetcher
func ProvideWebhooksUserIDFetcher() webhooks.UserIDFetcher {
	return UserIDFetcher
}

// ProvideWebhookIDFetcher provides an WebhookIDFetcher
func ProvideWebhookIDFetcher() webhooks.WebhookIDFetcher {
	return chiItemIDFetcher
}

// ProvideOAuth2ServiceClientIDFetcher provides a ClientIDFetcher
func ProvideOAuth2ServiceClientIDFetcher() oauth2clients.ClientIDFetcher {
	return chiOAuth2ClientIDFetcher
}

// UserIDFetcher fetches a user ID from a request routed by chi.
func UserIDFetcher(req *http.Request) uint64 {
	x, _ := req.Context().Value(models.UserIDKey).(uint64)
	return x
}

// ChiUserIDFetcher fetches a Username from a request routed by chi.
func ChiUserIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, users.URIParamKey), 10, 64)
	return u
}

// chiItemIDFetcher fetches a Username from a request routed by chi.
func chiItemIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, items.URIParamKey), 10, 64)
	return u
}

// chiOAuth2ClientIDFetcher fetches a Username from a request routed by chi.
func chiOAuth2ClientIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, oauth2clients.URIParamKey), 10, 64)
	return u
}
