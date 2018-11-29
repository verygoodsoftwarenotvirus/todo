package server

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/items"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	maxTimeout = 120 * time.Second
)

type ServerConfig struct {
	CertFile  string
	KeyFile   string
	DebugMode bool

	DBBuilder func(database.Config) (database.Database, error)
}

type Server struct {
	DebugMode bool
	certFile  string
	keyFile   string

	authenticator auth.Enticator
	db            database.Database

	// Services
	itemsService *items.ItemsService

	// infra things
	upgrader websocket.Upgrader
	server   *http.Server
	router   *chi.Mux
	logger   *logrus.Logger
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

func (s *Server) setupRoutes() *chi.Mux {
	router := setupRouter()

	router.Get("/_debug_/health", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	})

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {
			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := "/{itemID:[0-9]+}"
				itemsRouter.
					With(s.itemsService.ItemContextMiddleware).
					Post("/", s.itemsService.Create) // Create
				itemsRouter.Get(sr, s.itemsService.Read)      // Read
				itemsRouter.Get("/", s.itemsService.List)     // List
				itemsRouter.Delete(sr, s.itemsService.Delete) // Delete
				itemsRouter.
					With(s.itemsService.ItemContextMiddleware).
					Put(sr, s.itemsService.Update) // Update
			})
		})
	})

	return router
}

func NewServer(cfg ServerConfig, dbConfig database.Config) (*Server, error) {
	logger := logrus.New()

	dbConfig.Logger = logger
	db, err := cfg.DBBuilder(dbConfig)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(dbConfig.SchemaDir); err != nil {
		return nil, err
	}

	srv := &Server{
		DebugMode: cfg.DebugMode,
		db:        db,
		logger:    logger,
		certFile:  cfg.CertFile,
		keyFile:   cfg.KeyFile,
		server:    buildServer(),
		router:    setupRouter(),
		upgrader:  websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},

		// Services
		itemsService: items.NewItemsService(
			items.ItemsServiceConfig{
				Logger: logger,
				DB:     db,
			},
		),
	}
	srv.server.Handler = srv.setupRoutes()
	return srv, nil
}

func NewDebug(cfg ServerConfig, dbConfig database.Config) (*Server, error) {
	dbConfig.Debug = true
	cfg.DebugMode = true
	c, err := NewServer(cfg, dbConfig)
	if err != nil {
		return nil, err
	}
	c.logger.SetLevel(logrus.DebugLevel)
	return c, nil
}

func (s *Server) Serve() {
	s.logger.Debugf("Listening on 443. Go to https://localhost/")
	s.logger.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}

func parseEventMap(params url.Values) map[string]bool {
	out := map[string]bool{}
	if x, ok := params["events"]; ok && len(x) > 0 {
		for _, y := range x {
			z := strings.ToLower(y)
			if _, ok := models.ValidEventMap[z]; ok {
				out[models.ValidEventMap[z]] = true
			}
		}
	}
	return out
}

func parseTypeCollection(params url.Values) []string {
	out := []string{}
	if x, ok := params["events"]; ok && len(x) > 0 {
		if len(x) == 1 && x[0] == "*" {
			return models.AllEvents
		}

		for _, y := range x {
			switch y {
			case string(models.Create):
				out = append(out, y)
			case string(models.Update):
				out = append(out, y)
			case string(models.Delete):
				out = append(out, y)
			}
		}
	}
	return out
}
