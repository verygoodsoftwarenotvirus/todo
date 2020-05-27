package httpserver

import (
	"net/http"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
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
		ProvideUsersServiceUserIDFetcher,
		ProvideOAuth2ClientsServiceClientIDFetcher,
		ProvideItemsServiceUserIDFetcher,
		ProvideItemsServiceItemIDFetcher,
		ProvideWebhooksServiceUserIDFetcher,
		ProvideWebhooksServiceWebhookIDFetcher,
	)
)

// ProvideItemsServiceUserIDFetcher provides a UserIDFetcher.
func ProvideItemsServiceUserIDFetcher() itemsservice.UserIDFetcher {
	return userIDFetcherFromRequestContext
}

// ProvideItemsServiceItemIDFetcher provides an ItemIDFetcher.
func ProvideItemsServiceItemIDFetcher(logger logging.Logger) itemsservice.ItemIDFetcher {
	return buildRouteParamItemIDFetcher(logger)
}

// ProvideUsersServiceUserIDFetcher provides a UsernameFetcher.
func ProvideUsersServiceUserIDFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return buildRouteParamUserIDFetcher(logger)
}

// ProvideWebhooksServiceUserIDFetcher provides a UserIDFetcher.
func ProvideWebhooksServiceUserIDFetcher() webhooksservice.UserIDFetcher {
	return userIDFetcherFromRequestContext
}

// ProvideWebhooksServiceWebhookIDFetcher provides an WebhookIDFetcher.
func ProvideWebhooksServiceWebhookIDFetcher(logger logging.Logger) webhooksservice.WebhookIDFetcher {
	return buildRouteParamWebhookIDFetcher(logger)
}

// ProvideOAuth2ClientsServiceClientIDFetcher provides a ClientIDFetcher.
func ProvideOAuth2ClientsServiceClientIDFetcher(logger logging.Logger) oauth2clientsservice.ClientIDFetcher {
	return buildRouteParamOAuth2ClientIDFetcher(logger)
}

// userIDFetcherFromRequestContext fetches a user ID from a request routed by chi.
func userIDFetcherFromRequestContext(req *http.Request) uint64 {
	if si, ok := req.Context().Value(models.SessionInfoKey).(*models.SessionInfo); ok && si != nil {
		return si.UserID
	}
	return 0
}

// buildRouteParamUserIDFetcher builds a function that fetches a Username from a request routed by chi.
func buildRouteParamUserIDFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return func(req *http.Request) uint64 {
		u, err := strconv.ParseUint(chi.URLParam(req, usersservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching user ID from request")
		}
		return u
	}
}

// buildRouteParamItemIDFetcher builds a function that fetches a ItemID from a request routed by chi.
func buildRouteParamItemIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, itemsservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching ItemID from request")
		}
		return u
	}
}

// buildRouteParamWebhookIDFetcher fetches a WebhookID from a request routed by chi.
func buildRouteParamWebhookIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, webhooksservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching WebhookID from request")
		}
		return u
	}
}

// buildRouteParamOAuth2ClientIDFetcher fetches a OAuth2ClientID from a request routed by chi.
func buildRouteParamOAuth2ClientIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, oauth2clientsservice.URIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching OAuth2ClientID from request")
		}
		return u
	}
}
