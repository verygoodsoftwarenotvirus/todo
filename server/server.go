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
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/items"

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

	DBBuilder func(database.Config) (database.Database, error)
}

type Server struct {
	DebugMode bool
	certFile  string
	keyFile   string

	server        *http.Server
	authenticator auth.Enticator
	db            database.Database

	// Services
	itemsService *items.ItemsService

	// infra things
	router *chi.Mux
	logger *logrus.Logger
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
				s.logger.Errorf("error encountered decoding request body: %v", err)
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
			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := "/{itemID:[0-9]+}"
				// Create
				itemsRouter.
					With(s.buildRouteCtx(
						models.ItemInputCtxKey,
						new(models.ItemInput),
					)).
					Post("/", s.itemsService.Create)

				// Read
				itemsRouter.Get(sr, s.itemsService.Read)
				// List
				itemsRouter.Get("/", s.itemsService.List)

				// Update
				itemsRouter.
					With(s.buildRouteCtx(
						models.ItemInputCtxKey,
						new(models.ItemInput),
					)).
					Put(sr, s.itemsService.Update)

				// Delete
				itemsRouter.Delete(sr, s.itemsService.Delete)
			})
		})
	})

	return router
}

func NewServer(cfg ServerConfig, dbConfig database.Config) (*Server, error) {
	logger := logrus.New()

	db, err := cfg.DBBuilder(dbConfig)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(cfg.SchemaDir); err != nil {
		return nil, err
	}

	return &Server{
		certFile: cfg.CertFile,
		keyFile:  cfg.KeyFile,
		server:   buildServer(),
		logger:   logger,
		router:   setupRouter(),
		db:       db,

		itemsService: items.NewItemsService(items.ItemsServiceConfig{
			Logger: logger,
			DB:     db,
		}),
	}, nil
}

func NewDebug(cfg ServerConfig, dbConfig database.Config) (*Server, error) {
	c, err := NewServer(cfg, dbConfig)
	if err != nil {
		return nil, err
	}
	c.DebugMode = true
	c.logger.SetLevel(logrus.DebugLevel)
	return c, nil
}

func (s *Server) Serve() {
	s.server.Handler = s.setupRoutes()
	s.logger.Debugf("Listening on 443. Go to https://localhost/")
	s.logger.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}
