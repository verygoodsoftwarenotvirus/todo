package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime"
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
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/oauth2.v3"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// TodoServer is an obligatory interface
	TodoServer interface {
		Serve()
	}

	// CertPair represents the certificate and key you need to serve HTTPS
	CertPair struct {
		CertFile string
		KeyFile  string
	}

	// Tracer is an arbitrary type we use for dependency injection
	Tracer opentracing.Tracer

	// Server is our API server
	Server struct {
		DebugMode bool
		certFile  string
		keyFile   string

		authenticator auth.Enticator

		// Services
		itemsService         *items.Service
		usersService         *users.Service
		oauth2ClientsService *oauth2clients.Service

		// infra things
		db     database.Database
		router *chi.Mux
		server *http.Server
		logger logging.Logger
		tracer opentracing.Tracer

		// Auth stuff
		cookieBuilder     *securecookie.SecureCookie
		oauth2Handler     *oauth2server.Server
		oauth2ClientStore *oauth2store.ClientStore
	}
)

var (
	// Providers is our wire superset of providers this package offers
	Providers = wire.NewSet(
		paramFetcherProviders,
		ProvideTokenStore,
		ProvideClientStore,
		ProvideServer,
		ProvideHTTPServer,
		ProvideServerTracer,
		ProvideOAuth2Server,
	)
)

// ProvideMetricsNamespace provides a metrics namespace
func ProvideMetricsNamespace() metrics.Namespace {
	return "todo-server"
}

// ProvideServerTracer provides a UserServiceTracer from an tracer building function
func ProvideServerTracer() (Tracer, error) {
	return tracing.ProvideTracer("todo-server")
}

// ProvideServer builds a new Server instance
func ProvideServer(
	debug bool,
	cp CertPair,
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
	metricsHandler metrics.Handler,

	// OAuth2 stuff
	oauth2Handler *oauth2server.Server,
	tokenStore oauth2.TokenStore,
	clientStore *oauth2store.ClientStore,
) (*Server, error) {

	if len(cookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	span := tracer.StartSpan("startup")
	defer span.Finish()

	if err := db.Migrate(context.Background()); err != nil {
		return nil, err
	}

	srv := &Server{
		DebugMode:     debug,
		certFile:      cp.CertFile,
		keyFile:       cp.KeyFile,
		authenticator: authenticator,

		// infra thngs
		db:            db,
		logger:        logger,
		server:        server,
		cookieBuilder: securecookie.New(securecookie.GenerateRandomKey(64), cookieSecret),
		tracer:        tracer,

		// Services
		usersService:         usersService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,

		// OAuth2 stuff
		oauth2ClientStore: clientStore,
		oauth2Handler:     oauth2Handler,
	}

	srv.setupRoutes(metricsHandler)
	srv.initializeOAuth2Clients()

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.server.Handler = prometheus.InstrumentHandler("todo-server", s.router)
	s.logger.Debug("Listening on 443")
	log.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}

func (s *Server) stats(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("stats called")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	x := struct {
		Alloc      uint64 `json:"memory_allocated"`
		TotalAlloc uint64 `json:"lifetime_memory_allocated"`
		NumGC      uint64 `json:"num_gc"`
	}{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		NumGC:      uint64(m.NumGC),
	}

	if err := json.NewEncoder(res).Encode(x); err != nil {
		s.logger.Error(err, "encoding struct")
	}
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
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "Unexpected internal error occurred"})
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
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "Unauthorized"})
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
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid input"})
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
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "not found"})
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
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid cookie"})
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
