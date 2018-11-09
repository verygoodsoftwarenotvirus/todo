package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

const (
	maxTimeout = 120 * time.Second
)

type ServerConfig struct {
	CertFile  string
	KeyFile   string
	DebugMode bool
	SchemaDir string

	DBConfig  database.Config
	DBBuilder func(database.Config) (database.Database, error)
}

type Server struct {
	DebugMode bool
	certFile  string
	keyFile   string

	server        *http.Server
	Authenticator auth.Enticator
	db            database.Database

	Router *chi.Mux
	Logger *logrus.Logger
}

func buildServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  maxTimeout,
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

func setupRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.DefaultLogger)
	router.Use(middleware.Timeout(maxTimeout))

	return router
}

func (s *Server) buildRouteCtx(key models.ContextKey, x interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if err := json.NewDecoder(req.Body).Decode(x); err != nil {
				s.Logger.Errorf("error encountered decoding request body: %v", err)
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(req.Context(), key, x)
			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}

func (s *Server) setupRoutes() *chi.Mux {
	router := setupRouter()

	router.Get(
		"/_debug_/health",
		func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		},
	)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {
			// Items
			v1Router.
				With(s.buildRouteCtx(models.ItemInputCtxKey, new(models.ItemInput))).
				Post("/items", s.createItem) // Create

			v1Router.Get("/item/{itemID:[0-9]+}", s.getItem) // Read
			v1Router.Get("/items", s.getItems)               // List

			v1Router.
				With(s.buildRouteCtx(models.ItemInputCtxKey, new(models.ItemInput))).
				Patch("/item/{itemID:[0-9]+}", s.updateItem) // Update

			v1Router.Delete("/item/{itemID:[0-9]+}", s.deleteItem) // Delete
		})
	})

	return router
}

func NewServer(cfg ServerConfig) (*Server, error) {
	logger := logrus.New()

	db, err := cfg.DBBuilder(cfg.DBConfig)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(cfg.SchemaDir); err != nil {
		return nil, err
	}

	db.CreateItem(&models.ItemInput{
		Name:    "name",
		Details: "details",
	})

	db.CreateItem(&models.ItemInput{
		Name:    "other",
		Details: "things",
	})

	return &Server{
		certFile: cfg.CertFile,
		keyFile:  cfg.KeyFile,
		server:   buildServer(),
		Logger:   logger,
		Router:   setupRouter(),
		db:       db,
	}, nil
}

func NewDebug(cfg ServerConfig) (*Server, error) {
	c, err := NewServer(cfg)
	if err != nil {
		return nil, err
	}
	c.DebugMode = true
	c.Logger.SetLevel(logrus.DebugLevel)
	return c, nil
}

func (s *Server) Serve() {
	s.server.Handler = s.setupRoutes()
	s.Logger.Debugf("About to listen on 443. Go to https://localhost/")
	s.Logger.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}
