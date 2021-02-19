package routing

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// all interfaces HEAVILY inspired by github.com/go-chi/chi.

type (
	// Middleware  is a type alias for a middleware handler function.
	Middleware func(http.Handler) http.Handler

	// Router defines the contract between routing library and caller.
	Router interface {
		LogRoutes()
		Handler() http.Handler
		WithMiddleware(middleware ...Middleware) Router
		AddRoute(method, path string, handler http.HandlerFunc, middleware ...Middleware) error

		// Handle and HandleFunc adds routes for `pattern` that matches
		// all HTTP methods.
		Handle(pattern string, handler http.Handler)
		HandleFunc(pattern string, handler http.HandlerFunc)

		// Route mounts a sub-Router along a `pattern`` string.
		Route(pattern string, fn func(r Router)) Router

		// HTTP-method routing along `pattern`
		Connect(pattern string, handler http.HandlerFunc)
		Delete(pattern string, handler http.HandlerFunc)
		Get(pattern string, handler http.HandlerFunc)
		Head(pattern string, handler http.HandlerFunc)
		Options(pattern string, handler http.HandlerFunc)
		Patch(pattern string, handler http.HandlerFunc)
		Post(pattern string, handler http.HandlerFunc)
		Put(pattern string, handler http.HandlerFunc)
		Trace(pattern string, handler http.HandlerFunc)
	}

	// RouteParamManager builds route param fetchers for a compatible router.
	RouteParamManager interface {
		// Route params
		UserIDFetcherFromRequestContext(req *http.Request) uint64
		SessionInfoFetcherFromRequestContext(req *http.Request) (*types.RequestContext, error)
		BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(req *http.Request) uint64
		BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string
	}
)
