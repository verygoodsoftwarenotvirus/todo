package chi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-chi/chi"
)

type chirouteParamManager struct{}

var (
	errNoRequestContextAvailable = errors.New("no RequestContext attached to request context")
)

// NewRouteParamManager provides a new RouteParamManager.
func NewRouteParamManager() routing.RouteParamManager {
	return &chirouteParamManager{}
}

// UserIDFetcherFromRequestContext fetches a user ID from a request.
func (r chirouteParamManager) UserIDFetcherFromRequestContext(req *http.Request) uint64 {
	if reqCtx, err := r.FetchContextFromRequest(req); err == nil && reqCtx != nil {
		return reqCtx.User.ID
	}

	return 0
}

// requestContextFetcherFromRequestContext fetches a RequestContext from a request.
func (r chirouteParamManager) FetchContextFromRequest(req *http.Request) (*types.RequestContext, error) {
	if reqCtx, ok := req.Context().Value(types.RequestContextKey).(*types.RequestContext); ok && reqCtx != nil {
		return reqCtx, nil
	}

	return nil, errNoRequestContextAvailable
}

// BuildRouteParamIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func (r chirouteParamManager) BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(req *http.Request) uint64 {
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
func (r chirouteParamManager) BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string {
	return func(req *http.Request) string {
		return chi.URLParam(req, key)
	}
}
