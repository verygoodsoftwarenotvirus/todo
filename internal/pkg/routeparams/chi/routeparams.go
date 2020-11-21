package chi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-chi/chi"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// UserIDFetcherFromRequestContext fetches a user ID from a request routed by chi.
// NOTE: this function isn't technically a URI param fetcher, but it does fetch
// something from the request context, which is what chi.URLParam does too.
func UserIDFetcherFromRequestContext(req *http.Request) uint64 {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si.UserID
	}

	return 0
}

// SessionInfoFetcherFromRequestContext fetches a SessionInfo from a request routed by chi.
// NOTE: this function isn't technically a URI param fetcher, but it does fetch
// something from the request context, which is what chi.URLParam does too.
func SessionInfoFetcherFromRequestContext(req *http.Request) (*types.SessionInfo, error) {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si, nil
	}

	return nil, errors.New("no session info attached to request")
}

// BuildRouteParamIDFetcher builds a function that fetches a given key from a path with variables routed by chi.
func BuildRouteParamIDFetcher(logger logging.Logger, key, thingName string) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// we can generally disregard this error only because we should be able to validate.
		// that the string only contains numbers via chi's regex url param feature.
		u, err := strconv.ParseUint(chi.URLParam(req, key), 10, 64)
		if err != nil {
			logger.Error(err, fmt.Sprintf("fetching %s ID from request", thingName))
		}

		return u
	}
}
