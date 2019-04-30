package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
	//
	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// Server is our API httpServer
	Server struct {
		DebugMode bool

		// Services
		authService          *auth.Service
		itemsService         models.ItemDataServer
		usersService         models.UserDataServer
		oauth2ClientsService models.Oauth2ClientDataServer

		// infra things
		db         database.Database
		config     *config.ServerConfig
		router     *chi.Mux
		httpServer *http.Server
		logger     logging.Logger
		encoder    encoding.ServerEncoder

		// Auth stuff
		adminUserExists bool
		cookieBuilder   *securecookie.SecureCookie
	}
)

// ProvideServer builds a new Server instance
func ProvideServer(
	config *config.ServerConfig,

	// services
	authService *auth.Service,
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,

	// infra things
	db database.Database,
	logger logging.Logger,
	encoder encoding.EncoderDecoder,
) (*Server, error) {

	if len(config.Auth.CookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	srv := &Server{
		DebugMode: config.Server.Debug,

		// infra things
		db:            db,
		config:        config,
		encoder:       encoder,
		httpServer:    provideHTTPServer(),
		logger:        logger.WithName("api_server"),
		cookieBuilder: securecookie.New(securecookie.GenerateRandomKey(64), []byte(config.Auth.CookieSecret)),

		// services
		usersService:         usersService,
		authService:          authService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	// begin new monitoring stuff

	// Firstly, we'll register ochttp Server views.
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register server views for HTTP metrics: %v", err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: string(config.Meta.MetricsNamespace),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Prometheus exporter")
	}
	view.RegisterExporter(pe)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: fmt.Sprintf("%s:%s", os.Getenv("JAEGER_AGENT_HOST"), os.Getenv("JAEGER_AGENT_PORT")),
		Process: jaeger.Process{
			ServiceName: os.Getenv("JAEGER_SERVICE_NAME"),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Jaeger exporter")
	}

	trace.RegisterExporter(je)

	// end new monitoring stuff

	srv.setupRouter(config.Server.FrontendFilesDirectory, pe)
	srv.httpServer.Handler = &ochttp.Handler{
		Handler:        srv.router,
		FormatSpanName: formatSpanNameForRequest,
	}

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.httpServer.Addr = fmt.Sprintf(":%d", s.config.Server.HTTPPort)
	s.logger.Debug(fmt.Sprintf("Listening for HTTP requests on %s", s.httpServer.Addr))
	log.Fatal(s.httpServer.ListenAndServe())
}

type genericResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ErrorNotifier is a function which can notify a user of an error
type ErrorNotifier func(res http.ResponseWriter, req *http.Request, err error)

func (s *Server) internalServerError(res http.ResponseWriter, req *http.Request, err error) {
	s.logger.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Debug("internalServerError called")

	sc := http.StatusInternalServerError
	s.logger.Error(err, "Encountered internal error")
	res.WriteHeader(sc)

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "Unexpected internal error occurred"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func (s *Server) notifyUnauthorized(res http.ResponseWriter, req *http.Request, err error) {
	s.logger.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Debug("notifyUnauthorized called")

	sc := http.StatusUnauthorized
	if err != nil {
		s.logger.WithError(err).Debug("notifyUnauthorized called")
	}
	res.WriteHeader(sc)

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "Unauthorized"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func (s *Server) invalidInput(res http.ResponseWriter, req *http.Request, err error) {
	s.logger.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Debug("invalidInput called")

	sc := http.StatusBadRequest
	s.logger.WithValue("route", req.URL.Path).Debug("invalidInput called for route")
	if err != nil {
		s.logger.WithError(err).Debug("invalidInput called")
	}
	res.WriteHeader(sc)

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "invalid input"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func (s *Server) notFound(res http.ResponseWriter, req *http.Request, err error) {
	s.logger.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Debug("notFound called")

	sc := http.StatusNotFound
	s.logger.WithValue("route", req.URL.Path).Debug("notFound called for route")
	if err != nil {
		s.logger.WithError(err).Debug("notFound called")
	}
	res.WriteHeader(sc)

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "not found"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func (s *Server) notifyOfInvalidRequestCookie(res http.ResponseWriter, req *http.Request, err error) {
	s.logger.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Debug("notifyOfInvalidRequestCookie called")

	sc := http.StatusBadRequest
	s.logger.WithValue("route", req.URL.Path).Debug("notifyOfInvalidRequestCookie called for route")
	if err != nil {
		s.logger.WithError(err).Debug("notifyOfInvalidRequestCookie called")
	}
	res.WriteHeader(sc)

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "invalid cookie"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}
