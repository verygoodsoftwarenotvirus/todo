package httpserver

import (
	"net/http"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

var (
	paramFetcherProviders = wire.NewSet(
		ProvideItemServiceUserIDFetcher,
		ProvideUsernameFetcher,
		ProvideOAuth2ServiceClientIDFetcher,
		ProvideAuthUserIDFetcher,
		ProvideWebhooksUserIDFetcher,
		ProvideItemIDFetcher,
		ProvideWebhookIDFetcher,
	)
)

// ProvideItemServiceUserIDFetcher provides a UserIDFetcher
func ProvideItemServiceUserIDFetcher() itemsservice.UserIDFetcher {
	return UserIDFetcher
}

// ProvideItemIDFetcher provides an ItemIDFetcher
func ProvideItemIDFetcher(logger logging.Logger) itemsservice.ItemIDFetcher {
	return buildChiItemIDFetcher(logger)
}

// ProvideUsernameFetcher provides a UsernameFetcher
func ProvideUsernameFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return buildChiUserIDFetcher(logger)
}

// ProvideAuthUserIDFetcher provides a UsernameFetcher
func ProvideAuthUserIDFetcher() authservice.UserIDFetcher {
	return UserIDFetcher
}

// ProvideWebhooksUserIDFetcher provides a UserIDFetcher
func ProvideWebhooksUserIDFetcher() webhooksservice.UserIDFetcher {
	return UserIDFetcher
}

// ProvideWebhookIDFetcher provides an WebhookIDFetcher
func ProvideWebhookIDFetcher(logger logging.Logger) webhooksservice.WebhookIDFetcher {
	return buildChiWebhookIDFetcher(logger)
}

// ProvideOAuth2ServiceClientIDFetcher provides a ClientIDFetcher
func ProvideOAuth2ServiceClientIDFetcher(logger logging.Logger) oauth2clientsservice.ClientIDFetcher {
	return buildChiOAuth2ClientIDFetcher(logger)
}

// UserIDFetcher fetches a user ID from a request routed by chi.
func UserIDFetcher(req *http.Request) uint64 {
	if userID, ok := req.Context().Value(models.UserIDKey).(uint64); ok {
		return userID
	}
	return 0
}

// buildChiUserIDFetcher builds a function that fetches a Username from a request routed by chi.
func buildChiUserIDFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return func(req *http.Request) uint64 {
		u, err := strconv.ParseUint(chi.URLParam(req, usersservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching user ID from request")
		}
		return u
	}
}

// buildChiItemIDFetcher builds a function that fetches a ItemID from a request routed by chi.
func buildChiItemIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, itemsservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching ItemID from request")
		}
		return u
	}
}

// chiWebhookIDFetcher fetches a WebhookID from a request routed by chi.
func buildChiWebhookIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, webhooksservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching WebhookID from request")
		}
		return u
	}
}

// chiOAuth2ClientIDFetcher fetches a OAuth2ClientID from a request routed by chi.
func buildChiOAuth2ClientIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, oauth2clientsservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching OAuth2ClientID from request")
		}
		return u
	}
}
