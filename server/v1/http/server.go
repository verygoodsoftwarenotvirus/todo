package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/opentracing/opentracing-go"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// Tracer is an arbitrary type we use for dependency injection
	Tracer opentracing.Tracer

	// Server is our API httpServer
	Server struct {
		DebugMode bool

		authenticator auth.Enticator

		// Services
		itemsService         models.ItemDataServer
		usersService         models.UserDataServer
		oauth2ClientsService models.Oauth2ClientDataServer

		// infra things
		db         database.Database
		config     *config.ServerConfig
		router     *chi.Mux
		httpServer *http.Server
		logger     logging.Logger
		tracer     opentracing.Tracer
		encoder    encoding.ServerEncoderDecoder

		// Auth stuff
		adminUserExists bool
		cookieBuilder   *securecookie.SecureCookie
	}
)

// ProvideServer builds a new Server instance
func ProvideServer(
	config *config.ServerConfig,
	authenticator auth.Enticator,

	// services
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,

	// infra things
	db database.Database,
	logger logging.Logger,
	httpServer *http.Server,
	encoder encoding.ServerEncoderDecoder,

	// metrics things
	metricsHandler metrics.Handler,
	metricsMiddleware metrics.Middleware,
) (*Server, error) {

	if len(config.Auth.CookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	cookieBuilder := securecookie.New(securecookie.GenerateRandomKey(64), []byte(config.Auth.CookieSecret))

	srv := &Server{
		DebugMode:     config.Server.Debug,
		authenticator: authenticator,

		// infra things
		db:            db,
		config:        config,
		logger:        logger.WithName("httperver"),
		httpServer:    httpServer,
		cookieBuilder: cookieBuilder,
		tracer:        tracing.ProvideTracer("httperver"),
		encoder:       encoder,

		// services
		usersService:         usersService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	srv.logger.Info("migrating database")
	if err := srv.db.Migrate(context.Background()); err != nil {
		return nil, err
	}
	srv.logger.Info("database migrated!")

	srv.setupRouter(config.Server.FrontendFilesDirectory, metricsHandler, metricsMiddleware)
	srv.httpServer.Handler = srv.router

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
