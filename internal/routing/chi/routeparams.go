package chi

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"

	"github.com/go-chi/chi"
)

type chiRouteParamManager struct{}

// NewRouteParamManager provides a new RouteParamManager.
func NewRouteParamManager() routing.RouteParamManager {
	return &chiRouteParamManager{}
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
