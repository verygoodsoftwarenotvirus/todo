package server

import (
	"context"
	"crypto/tls"
	"errors"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"log"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/google/wire"
	"github.com/gorilla/securecookie"
	"github.com/opentracing/opentracing-go"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// TodoServer is an obligatory interface
	TodoServer interface {
		Serve()
	}

	// Tracer is an arbitrary type we use for dependency injection
	Tracer opentracing.Tracer

	// Server is our API server
	Server struct {
		DebugMode bool

		authenticator auth.Enticator

		// Services
		itemsService         *items.Service
		usersService         *users.Service
		oauth2ClientsService *oauth2clients.Service

		// infra things
		db      database.Database
		router  *chi.Mux
		server  *http.Server
		logger  logging.Logger
		tracer  opentracing.Tracer
		encoder encoding.ResponseEncoder

		// Auth stuff
		cookieBuilder *securecookie.SecureCookie
	}
)

var (
	// Providers is our wire superset of providers this package offers
	Providers = wire.NewSet(
		paramFetcherProviders,
		ProvideServer,
		ProvideHTTPServer,
		ProvideServerTracer,
	)
)

// ProvideServerTracer provides a UserServiceTracer from an tracer building function
func ProvideServerTracer() (Tracer, error) {
	return tracing.ProvideTracer("todo-server")
}

// ProvideServer builds a new Server instance
func ProvideServer(
	debug bool,
	cookieSecret []byte,
	authenticator auth.Enticator,

	// services
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,

	// infra things
	db database.Database,
	logger logging.Logger,
	tracer Tracer,
	server *http.Server,
	encoder encoding.ResponseEncoder,

	// metrics things
	metricsHandler metrics.Handler,
	instHandlerProvider metrics.InstrumentationHandlerProvider,
) (*Server, error) {

	if len(cookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	cookieBuilder := securecookie.New(securecookie.GenerateRandomKey(64), cookieSecret)

	srv := &Server{
		DebugMode:     debug,
		authenticator: authenticator,

		// infra thngs
		db:            db,
		logger:        logger,
		server:        server,
		cookieBuilder: cookieBuilder,
		tracer:        tracer,
		encoder:       encoder,

		// Services
		usersService:         usersService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	srv.logger.Info("migrating database")
	if err := srv.db.Migrate(context.Background()); err != nil {
		return nil, err
	}
	srv.logger.Info("database migrated!")

	cc := srv.oauth2ClientsService.InitializeOAuth2Clients()
	srv.setupRouter(metricsHandler, cc == 0)

	var handler http.Handler = srv.router
	if instHandlerProvider != nil {
		srv.logger.Debug("setting instrumentation handler")
		handler = instHandlerProvider(srv.router)
	}
	srv.server.Handler = handler
	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.server.Addr = ":80"
	s.logger.Debug("Listening on 80")
	log.Fatal(s.server.ListenAndServe())
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

// ProvideHTTPServer provides an HTTP server
func ProvideHTTPServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			// "Only use curves which have assembly implementations"
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}
	return srv
}
