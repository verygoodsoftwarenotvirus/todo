package server

import (
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
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
func ProvideUsernameFetcher() users.UsernameFetcher {
	return ChiUsernameFetcher
}

// UserIDFetcher fetches a user ID from a request routed by chi.
func UserIDFetcher(req *http.Request) uint64 {
	x, _ := req.Context().Value(models.UserIDKey).(uint64)
	return x
}

// ChiUsernameFetcher fetches a username from a request routed by chi.
func ChiUsernameFetcher(req *http.Request) string {
	return chi.URLParam(req, users.URIParamKey)
}

// chiUserIDFetcher fetches a username from a request routed by chi.
func chiUserIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, users.URIParamKey), 10, 64)
	return u
}

// chiItemIDFetcher fetches a username from a request routed by chi.
func chiItemIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, items.URIParamKey), 10, 64)
	return u
}

// chiOAuth2ClientIDFetcher fetches a username from a request routed by chi.
func chiOAuth2ClientIDFetcher(req *http.Request) string {
	// PONDER: if the only time we use users.URIParamKey is externally to the users package
	// does it really need to belong there?
	return chi.URLParam(req, oauth2clients.URIParamKey)
}

// chiUserIDFetcher fetches a username from a request routed by chi.
func chiOAuth2ClientDBIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, oauth2clients.URIParamKey), 10, 64)
	return u
}
