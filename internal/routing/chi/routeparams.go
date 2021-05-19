package chi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/go-chi/chi"
)

type chiRouteParamManager struct{}

var (
	errNoSessionContextDataAvailable = errors.New("no SessionContextData attached to session context data")
)

// NewRouteParamManager provides a new RouteParamManager.
func NewRouteParamManager() routing.RouteParamManager {
	return &chiRouteParamManager{}
}

// UserIDFetcherFromSessionContextData fetches a user ID from a request.
func (r chiRouteParamManager) UserIDFetcherFromSessionContextData(req *http.Request) uint64 {
	if sessionCtxData, err := r.FetchContextFromRequest(req); err == nil && sessionCtxData != nil {
		return sessionCtxData.Requester.UserID
	}

	return 0
}

// FetchContextFromRequest fetches a SessionContextData from a request.
func (r chiRouteParamManager) FetchContextFromRequest(req *http.Request) (*types.SessionContextData, error) {
	if sessionCtxData, ok := req.Context().Value(types.SessionContextDataKey).(*types.SessionContextData); ok && sessionCtxData != nil {
		return sessionCtxData, nil
	}

	return nil, errNoSessionContextDataAvailable
}

// BuildRouteParamIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func (r chiRouteParamManager) BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		// this should never happen
		u, err := strconv.ParseUint(chi.URLParam(req, key), 10, 64)
		if err != nil && len(logDescription) > 0 {
			logger.Error(err, fmt.Sprintf("fetching %s ID from request", logDescription))
		}

		return u
	}
}

// BuildRouteParamStringIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func (r chiRouteParamManager) BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string {
	return func(req *http.Request) string {
		return chi.URLParam(req, key)
	}
}
