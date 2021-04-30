package chi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/unrolled/secure"
)

const (
	roughlyOneYear = time.Hour * 24 * 365
	maxTimeout     = 120 * time.Second
	maxCORSAge     = 300
)

var (
	errInvalidMethod = errors.New("invalid method")
)

var _ routing.Router = (*router)(nil)

type router struct {
	router chi.Router
	tracer tracing.Tracer
	logger logging.Logger
}

func buildChiMux(logger logging.Logger) chi.Router {
	ch := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts,
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"RawHTML-Provider",
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           maxCORSAge,
	})

	s := secure.New(secure.Options{
		AllowedHosts:            []string{""},
		AllowedHostsAreRegex:    false,
		HostsProxyHeaders:       []string{"X-Forwarded-Hosts"},
		SSLRedirect:             true,
		SSLTemporaryRedirect:    false,
		SSLHost:                 "",
		SSLHostFunc:             nil,
		SSLProxyHeaders:         map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:              int64(roughlyOneYear.Seconds()),
		STSIncludeSubdomains:    true,
		STSPreload:              true,
		ForceSTSHeader:          false,
		FrameDeny:               true,
		CustomFrameOptionsValue: "",
		ContentTypeNosniff:      true,
		BrowserXssFilter:        true,
		CustomBrowserXssValue:   "",
		ContentSecurityPolicy:   "",
		PublicKey:               "",
		ReferrerPolicy:          "",
		FeaturePolicy:           "",
		ExpectCTHeader:          "",
		SecureContextKey:        "secureContext",
		IsDevelopment:           true,
	})

	mux := chi.NewRouter()
	mux.Use(
		s.Handler,
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		chimiddleware.Timeout(maxTimeout),
		logging.BuildLoggingMiddleware(logging.EnsureLogger(logger).WithName("router")),
		ch.Handler,
	)

	// all middleware must be defined before routes on a mux.

	return mux
}

func buildRouter(mux chi.Router, logger logging.Logger) *router {
	logger = logging.EnsureLogger(logger)

	if mux == nil {
		logger.Info("starting with a new mux")
		mux = buildChiMux(logger)
	}

	r := &router{
		router: mux,
		tracer: tracing.NewTracer("router"),
		logger: logger,
	}

	return r
}

func convertMiddleware(in ...routing.Middleware) []func(handler http.Handler) http.Handler {
	out := []func(handler http.Handler) http.Handler{}

	for _, x := range in {
		if x != nil {
			out = append(out, x)
		}
	}

	return out
}

// NewRouter constructs a new router.
func NewRouter(logger logging.Logger) routing.Router {
	return buildRouter(nil, logger)
}

func (r *router) clone() *router {
	return buildRouter(r.router, r.logger)
}

// WithMiddleware returns a router with certain middleware applied.
func (r *router) WithMiddleware(middleware ...routing.Middleware) routing.Router {
	x := r.clone()

	x.router = x.router.With(convertMiddleware(middleware...)...)

	return x
}

// LogRoutes logs the described routes.
func (r *router) LogRoutes() {
	if err := chi.Walk(r.router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		r.logger.WithValues(map[string]interface{}{
			"method": method,
			"route":  route,
		}).Debug("route found")

		return nil
	}); err != nil {
		r.logger.Error(err, "logging routes")
	}
}

// Route lets you apply a set of routes to a subrouter with a provided pattern.
func (r *router) Route(pattern string, fn func(r routing.Router)) routing.Router {
	r.router.Route(pattern, func(subrouter chi.Router) {
		fn(buildRouter(subrouter, r.logger))
	})

	return r
}

// AddRoute adds a route to the router.
func (r *router) AddRoute(method, path string, handler http.HandlerFunc, middleware ...routing.Middleware) error {
	switch strings.TrimSpace(strings.ToUpper(method)) {
	case http.MethodGet:
		r.router.With(convertMiddleware(middleware...)...).Get(path, handler)
	case http.MethodHead:
		r.router.With(convertMiddleware(middleware...)...).Head(path, handler)
	case http.MethodPost:
		r.router.With(convertMiddleware(middleware...)...).Post(path, handler)
	case http.MethodPut:
		r.router.With(convertMiddleware(middleware...)...).Put(path, handler)
	case http.MethodPatch:
		r.router.With(convertMiddleware(middleware...)...).Patch(path, handler)
	case http.MethodDelete:
		r.router.With(convertMiddleware(middleware...)...).Delete(path, handler)
	case http.MethodConnect:
		r.router.With(convertMiddleware(middleware...)...).Connect(path, handler)
	case http.MethodOptions:
		r.router.With(convertMiddleware(middleware...)...).Options(path, handler)
	case http.MethodTrace:
		r.router.With(convertMiddleware(middleware...)...).Trace(path, handler)
	default:
		return fmt.Errorf("%s: %w", method, errInvalidMethod)
	}

	return nil
}

// Handler our interface by wrapping the underlying router's Handler method.
func (r *router) Handler() http.Handler {
	return r.router
}

// Handle our interface by wrapping the underlying router's Handle method.
func (r *router) Handle(pattern string, handler http.Handler) {
	r.router.Handle(pattern, handler)
}

// HandleFunc satisfies our interface by wrapping the underlying router's HandleFunc method.
func (r *router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.router.HandleFunc(pattern, handler)
}

// Connect satisfies our interface by wrapping the underlying router's Connect method.
func (r *router) Connect(pattern string, handler http.HandlerFunc) {
	r.router.Connect(pattern, handler)
}

// Delete satisfies our interface by wrapping the underlying router's Delete method.
func (r *router) Delete(pattern string, handler http.HandlerFunc) {
	r.router.Delete(pattern, handler)
}

// Get satisfies our interface by wrapping the underlying router's Get method.
func (r *router) Get(pattern string, handler http.HandlerFunc) {
	r.router.Get(pattern, handler)
}

// Head satisfies our interface by wrapping the underlying router's Head method.
func (r *router) Head(pattern string, handler http.HandlerFunc) {
	r.router.Head(pattern, handler)
}

// Options satisfies our interface by wrapping the underlying router's Options method.
func (r *router) Options(pattern string, handler http.HandlerFunc) {
	r.router.Options(pattern, handler)
}

// Patch satisfies our interface by wrapping the underlying router's Patch method.
func (r *router) Patch(pattern string, handler http.HandlerFunc) {
	r.router.Patch(pattern, handler)
}

// Post satisfies our interface by wrapping the underlying router's Post method.
func (r *router) Post(pattern string, handler http.HandlerFunc) {
	r.router.Post(pattern, handler)
}

// Put satisfies our interface by wrapping the underlying router's Put method.
func (r *router) Put(pattern string, handler http.HandlerFunc) {
	r.router.Put(pattern, handler)
}

// Trace satisfies our interface by wrapping the underlying router's Trace method.
func (r *router) Trace(pattern string, handler http.HandlerFunc) {
	r.router.Trace(pattern, handler)
}

// BuildRouteParamIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func (r *router) BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(req *http.Request) uint64 {
	return func(req *http.Request) uint64 {
		v := chi.URLParam(req, key)
		u, err := strconv.ParseUint(v, 10, 64)
		// this should never happen
		if err != nil && len(logDescription) > 0 {
			logger.Error(err, fmt.Sprintf("fetching %s ID from request", logDescription))
		}

		return u
	}
}

// BuildRouteParamStringIDFetcher builds a function that fetches a given key from a path with variables added by a router.
func (r *router) BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string {
	return func(req *http.Request) string {
		return chi.URLParam(req, key)
	}
}
