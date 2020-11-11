package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var paramFetcherProviders = wire.NewSet(
	ProvideUsersServiceUserIDFetcher,
	ProvideUsersServiceSessionInfoFetcher,
	ProvideOAuth2ClientsServiceClientIDFetcher,
	ProvideWebhooksServiceWebhookIDFetcher,
	ProvideWebhooksServiceUserIDFetcher,
	ProvideWebhooksServiceSessionInfoFetcher,
	ProvideItemsServiceItemIDFetcher,
	ProvideItemsServiceSessionInfoFetcher,
	ProvideAuditServiceItemIDFetcher,
	ProvideAuditServiceSessionInfoFetcher,
	ProvideAuthServiceSessionInfoFetcher,
)

// ProvideUsersServiceUserIDFetcher provides a UsernameFetcher.
func ProvideUsersServiceUserIDFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return buildRouteParamUserIDFetcher(logger)
}

// ProvideUsersServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideUsersServiceSessionInfoFetcher() usersservice.SessionInfoFetcher {
	return sessionInfoFetcherFromRequestContext
}

// ProvideOAuth2ClientsServiceClientIDFetcher provides a ClientIDFetcher.
func ProvideOAuth2ClientsServiceClientIDFetcher(logger logging.Logger) oauth2clientsservice.ClientIDFetcher {
	return buildRouteParamOAuth2ClientIDFetcher(logger)
}

// ProvideWebhooksServiceWebhookIDFetcher provides an WebhookIDFetcher.
func ProvideWebhooksServiceWebhookIDFetcher(logger logging.Logger) webhooksservice.WebhookIDFetcher {
	return buildRouteParamWebhookIDFetcher(logger)
}

// ProvideWebhooksServiceUserIDFetcher provides a UserIDFetcher.
func ProvideWebhooksServiceUserIDFetcher() webhooksservice.UserIDFetcher {
	return userIDFetcherFromRequestContext
}

// ProvideWebhooksServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideWebhooksServiceSessionInfoFetcher() webhooksservice.SessionInfoFetcher {
	return sessionInfoFetcherFromRequestContext
}

// ProvideItemsServiceItemIDFetcher provides an ItemIDFetcher.
func ProvideItemsServiceItemIDFetcher(logger logging.Logger) itemsservice.ItemIDFetcher {
	return buildRouteParamItemIDFetcher(logger)
}

// ProvideItemsServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideItemsServiceSessionInfoFetcher() itemsservice.SessionInfoFetcher {
	return sessionInfoFetcherFromRequestContext
}

// ProvideAuditServiceItemIDFetcher provides an EntryIDFetcher.
func ProvideAuditServiceItemIDFetcher(logger logging.Logger) auditservice.EntryIDFetcher {
	return buildRouteParamEntryIDFetcher(logger)
}

// ProvideAuditServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuditServiceSessionInfoFetcher() auditservice.SessionInfoFetcher {
	return sessionInfoFetcherFromRequestContext
}

// ProvideAuthServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuthServiceSessionInfoFetcher() authservice.SessionInfoFetcher {
	return sessionInfoFetcherFromRequestContext
}

// userIDFetcherFromRequestContext fetches a user ID from a request routed by chi.
// NOTE: this function isn't technically a URI param fetcher, but it does fetch
// something from the request context, which is what chi.URLParam does too.
func userIDFetcherFromRequestContext(req *http.Request) uint64 {
	if si, ok := req.Context().Value(models.SessionInfoKey).(*models.SessionInfo); ok && si != nil {
		return si.UserID
	}

	return 0
}

var errNoSessionInfoAttachedToRequest = errors.New("no session info attached to request")

// sessionInfoFetcherFromRequestContext fetches a SessionInfo from a request routed by chi.
// NOTE: this function isn't technically a URI param fetcher, but it does fetch
// something from the request context, which is what chi.URLParam does too.
func sessionInfoFetcherFromRequestContext(req *http.Request) (*models.SessionInfo, error) {
	if si, ok := req.Context().Value(models.SessionInfoKey).(*models.SessionInfo); ok && si != nil {
		return si, nil
	}

	return nil, errNoSessionInfoAttachedToRequest
}

// buildRouteParamUserIDFetcher builds a function that fetches a EnsureUsername from a request routed by chi.
func buildRouteParamUserIDFetcher(logger logging.Logger) usersservice.UserIDFetcher {
	return func(req *http.Request) uint64 {
		u, err := strconv.ParseUint(chi.URLParam(req, usersservice.UserIDURIParamKey), 10, 64)
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
		u, err := strconv.ParseUint(chi.URLParam(req, itemsservice.ItemIDURIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching item ID from request")
		}

		return u
	}
}

// buildRouteParamEntryIDFetcher builds a function that fetches a ItemID from a request routed by chi.
func buildRouteParamEntryIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, auditservice.LogEntryURIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching audit log entry ID from request")
		}

		return u
	}
}

// buildRouteParamWebhookIDFetcher fetches a WebhookID from a request routed by chi.
func buildRouteParamWebhookIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, webhooksservice.WebhookIDURIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching webhook ID from request")
		}

		return u
	}
}

// buildRouteParamOAuth2ClientIDFetcher fetches a OAuth2ClientID from a request routed by chi.
func buildRouteParamOAuth2ClientIDFetcher(logger logging.Logger) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, oauth2clientsservice.OAuth2ClientIDURIParamKey), 10, 64)
		if err != nil {
			logger.Error(err, "fetching OAuth2 client ID from request")
		}

		return u
	}
}
