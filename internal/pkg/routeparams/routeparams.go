package routeparams

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-chi/chi"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var routeParamRetrievalFunc = chi.URLParam

// UserIDFetcherFromRequestContext fetches a user ID from a request.
func UserIDFetcherFromRequestContext(req *http.Request) uint64 {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si.UserID
	}

	return 0
}

// SessionInfoFetcherFromRequestContext fetches a SessionInfo from a request.
func SessionInfoFetcherFromRequestContext(req *http.Request) (*types.SessionInfo, error) {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si, nil
	}

	return nil, errors.New("no session info attached to request")
}

// BuildRouteParamIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// this should never happen if routeParamRetrievalFunc is correct
		u, err := strconv.ParseUint(routeParamRetrievalFunc(req, key), 10, 64)
		if err != nil && len(logDescription) > 0 {
			logger.Error(err, fmt.Sprintf("fetching %s ID from request", logDescription))
		}

		return u
	}
}
