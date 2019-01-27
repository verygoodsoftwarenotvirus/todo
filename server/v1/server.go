package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	maxTimeout = 120 * time.Second
)

// Config describes the configuration of a server
type Config struct {
	CertFile     string
	KeyFile      string
	DebugMode    bool
	CookieSecret []byte
	Logger       *logrus.Logger

	Authenticator auth.Enticator

	Tracer opentracing.Tracer

	DBBuilder func(database.Config) (database.Database, error)
}

// Server is our API server
type Server struct {
	DebugMode bool
	certFile  string
	keyFile   string

	authenticator auth.Enticator

	// Services
	itemsService         *items.ItemsService
	usersService         *users.UsersService
	oauth2ClientsService *oauth2clients.Oauth2ClientsService

	// infra things
	db            database.Database
	router        *chi.Mux
	server        *http.Server
	logger        *logrus.Logger
	tracer        opentracing.Tracer
	cookieBuilder *securecookie.SecureCookie

	// Oauth2 stuff
	oauth2Handler     *oauth2server.Server
	oauth2ClientStore *oauth2store.ClientStore
	// oauth2TokenStore  oauth2store.TokenStore
}

// DefaultServerConfig provides the default Config object
func DefaultServerConfig() *Config {
	logger := logrus.New()
	return &Config{
		Logger:        logger,
		Authenticator: auth.NewBcrypt(logger),
	}
}

// NewServer builds a new Server instance
func NewServer(cfg Config, dbConfig database.Config) (*Server, error) {
	var logger = cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	if dbConfig.Logger == nil {
		dbConfig.Logger = logger
	}

	if cfg.Authenticator == nil {
		cfg.Authenticator = auth.NewBcrypt(logger)
	}

	if len(cfg.CookieSecret) < 32 {
		logger.Errorln("cookie secret is too short, must be at least 32 characters in length")
		return nil, errors.New("cookie secret is too short, must be at least 32 characters in length")
	}

	db, err := cfg.DBBuilder(dbConfig)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(dbConfig.SchemaDir); err != nil {
		return nil, err
	}

	srv := &Server{
		DebugMode:     cfg.DebugMode,
		db:            db,
		logger:        logger,
		certFile:      cfg.CertFile,
		keyFile:       cfg.KeyFile,
		server:        buildServer(),
		authenticator: cfg.Authenticator,
		cookieBuilder: securecookie.New(securecookie.GenerateRandomKey(64), cfg.CookieSecret),

		tracer: cfg.Tracer,

		// Services
		usersService: users.NewUsersService(
			users.UsersServiceConfig{
				// CookieName: s.config.CookieName
				Logger:          logger,
				Database:        db,
				Authenticator:   cfg.Authenticator,
				UsernameFetcher: chiUsernameFetcher,
			},
		),
	}

	if srv.itemsService, err = items.NewItemsService(
		items.ItemsServiceConfig{
			Logger:        logger,
			Database:      db,
			UserIDFetcher: srv.userIDFetcher,
		},
	); err != nil {
		return nil, err
	}

	srv.initializeOauth2Server()
	srv.setupRoutes()

	return srv, nil
}

// NewDebug builds a new debug Server
func NewDebug(cfg Config, dbConfig database.Config) (srv *Server, err error) {
	dbConfig.Debug, cfg.DebugMode = true, true
	if srv, err = NewServer(cfg, dbConfig); err != nil {
		return nil, err
	}
	srv.logger.SetLevel(logrus.DebugLevel)
	return
}

func (s *Server) logRoutes(routes chi.Routes) {
	if s.DebugMode {
		for _, route := range routes.Routes() {
			s.logRoute("", route)
		}
	}
}

func (s *Server) logRoute(prefix string, route chi.Route) {
	rp := route.Pattern
	if prefix != "" {
		rp = prefix + rp
	}
	s.logger.Debugln(rp)
	if route.SubRoutes != nil {
		for _, sr := range route.SubRoutes.Routes() {
			s.logRoute(rp, sr)
		}
	}
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.server.Handler = nethttp.Middleware(s.tracer, s.router)
	s.logger.Debugf("Listening on 443")
	//s.logRoutes(s.router)
	s.logger.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}

func (s *Server) stats(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("stats called")
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
		s.logger.Errorf("Error encoding struct: %v", err)
	}
}

type genericResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ErrorNotifier is a function which can notify a user of an error
type ErrorNotifier func(res http.ResponseWriter, req *http.Request, err error)

func (s *Server) internalServerError(res http.ResponseWriter, req *http.Request, err error) {
	sc := http.StatusInternalServerError
	s.logger.Errorf("Encountered this internal error: %v\n", err)
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "Unexpected internal error occurred"})
}

func (s *Server) notifyUnauthorized(res http.ResponseWriter, req *http.Request, err error) {
	sc := http.StatusUnauthorized
	if err != nil {
		s.logger.Errorln("notifyUnauthorized called with this error: ", err)
	}
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "Unauthorized"})
}

func (s *Server) invalidInput(res http.ResponseWriter, req *http.Request, err error) {
	sc := http.StatusBadRequest
	s.logger.Debugf("invalidInput called for route: %q\n", req.URL.String())
	if err != nil {
		s.logger.Errorln("invalidInput called with this error: ", err)
	}
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid input"})
}

func (s *Server) notFound(res http.ResponseWriter, req *http.Request, err error) {
	sc := http.StatusNotFound
	s.logger.Debugf("notFound called for route: %q\n", req.URL.String())
	if err != nil {
		s.logger.Errorln("notFound called with this error: ", err)
	}
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "not found"})
}

func (s *Server) notifyOfInvalidRequestCookie(res http.ResponseWriter, req *http.Request, err error) {
	sc := http.StatusBadRequest
	s.logger.Debugf("notifyOfInvalidRequestCookie called for route: %q\n", req.URL.String())
	if err != nil {
		s.logger.Errorln("notifyOfInvalidRequestCookie called with this error: ", err)
	}
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid cookie"})
}

func buildServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
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
}
