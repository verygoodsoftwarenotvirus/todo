package server

import (
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
)

// chiUsernameFetcher fetches a username from a request routed by chi.
func chiUsernameFetcher(req *http.Request) string {
	// PONDER: if the only time we use users.URIParamKey is externally to the users package
	// does it really need to belong there?
	return chi.URLParam(req, users.URIParamKey)
}

// chiUserIDFetcher fetches a username from a request routed by chi.
func chiUserIDFetcher(req *http.Request) uint64 {
	// we disregard this error only because we're able to validate that the string only
	// contains numbers via chi's regex things
	u, _ := strconv.ParseUint(chi.URLParam(req, users.URIParamKey), 10, 64)
	return u
}
