package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2manage "gopkg.in/oauth2.v3/manage"
	oauth2models "gopkg.in/oauth2.v3/models"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	maxTimeout = 120 * time.Second
)

type ServerConfig struct {
	CertFile     string
	KeyFile      string
	DebugMode    bool
	CookieSecret []byte

	DBBuilder func(database.Config) (database.Database, error)
}

type Server struct {
	DebugMode bool
	certFile  string
	keyFile   string

	authenticator auth.Enticator

	// Services
	loginMonitor LoginMonitor
	itemsService *items.ItemsService
	usersService *users.UsersService

	// infra things
	db            database.Database
	server        *http.Server
	logger        *logrus.Logger
	cookieSecret  []byte
	cookieBuilder *securecookie.SecureCookie

	// Oauth2 stuff
	oauth2Handler *oauth2server.Server
}

func NewServer(cfg ServerConfig, dbConfig database.Config) (*Server, error) {
	logger := logrus.New()
	dbConfig.Logger = logger

	if len(cfg.CookieSecret) < 32 {
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
		loginMonitor:  &NoopLoginMonitor{}, // TODO: this makes me sad
		authenticator: auth.NewBcrypt(logger),
		cookieSecret:  cfg.CookieSecret,
		cookieBuilder: securecookie.New(securecookie.GenerateRandomKey(64), cfg.CookieSecret),

		// Services
		itemsService: items.NewItemsService(
			items.ItemsServiceConfig{
				Logger:   logger,
				Database: db,
			},
		),
		usersService: users.NewUsersService(
			users.UsersServiceConfig{
				Logger:   logger,
				Database: db,
			},
		),
	}

	srv.initializeOauth2Routes()
	srv.setupRoutes()

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
	s.logger.Debugf("Listening on 443")
	s.logger.Fatal(s.server.ListenAndServeTLS(s.certFile, s.keyFile))
}

func (s *Server) initializeOauth2Routes() {
	manager := oauth2manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(oauth2store.NewMemoryTokenStore())

	// client memory store
	clientStore := oauth2store.NewClientStore()
	clientStore.Set("000000", &oauth2models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "https://localhost",
	})
	manager.MapClientStorage(clientStore)

	s.oauth2Handler = oauth2server.NewDefaultServer(manager)
	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)

	s.oauth2Handler.SetInternalErrorHandler(func(err error) (re *oauth2errors.Response) {
		s.logger.Errorf("Internal Error: %v", err.Error())
		return
	})

	s.oauth2Handler.SetResponseErrorHandler(func(re *oauth2errors.Response) {
		s.logger.Errorf("Response Error: %v", re.Error)
	})

}

func (s *Server) setupRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(context.ClearHandler)
	router.Use(middleware.DefaultLogger)
	router.Use(middleware.Timeout(maxTimeout))
	router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	})

	if s.DebugMode {
		router.Get("/_debug_/stats", s.stats)
	}

	router.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.usersService.UserInputContextMiddleware).Post("/login", s.Login)
		userRouter.Post("/logout", s.Logout)
	})

	router.Route("/oauth2", func(oauthRouter chi.Router) {
		oauthRouter.Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2Handler.HandleAuthorizeRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
		})

		oauthRouter.Get("/token", func(res http.ResponseWriter, req *http.Request) {
			s.oauth2Handler.HandleTokenRequest(res, req)
		})
	})

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {

			v1Router.With(s.AuthorizationMiddleware).Post("/fart", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusTeapot)
			})

			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
				itemsRouter.With(s.itemsService.ItemContextMiddleware).Post("/", s.itemsService.Create) // Create
				itemsRouter.Get(sr, s.itemsService.Read)                                                // Read
				itemsRouter.Get("/", s.itemsService.List)                                               // List
				itemsRouter.Delete(sr, s.itemsService.Delete)                                           // Delete
				itemsRouter.With(s.itemsService.ItemContextMiddleware).Put(sr, s.itemsService.Update)   // Update
			})
		})
	})

	s.server.Handler = router
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

func (s *Server) internalServerError(res http.ResponseWriter, err error) {
	sc := http.StatusInternalServerError
	s.logger.Errorf("Encountered this error: %v\n", err)
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "Unexpected internal error occurred"})
}

func (s *Server) invalidInput(res http.ResponseWriter, req *http.Request) {
	sc := http.StatusBadRequest
	s.logger.Debugf("invalidInput called for route: %q\n", req.URL.String())
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid input"})
}

func (s *Server) notFound(res http.ResponseWriter, req *http.Request) {
	sc := http.StatusNotFound
	s.logger.Debugf("notFound called for route: %q\n", req.URL.String())
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "not found"})
}

func (s *Server) notifyOfInvalidRequestCookie(res http.ResponseWriter, req *http.Request) {
	sc := http.StatusBadRequest
	s.logger.Debugf("notifyOfInvalidRequestCookie called for route: %q\n", req.URL.String())
	res.WriteHeader(sc)
	json.NewEncoder(res).Encode(genericResponse{Status: sc, Message: "invalid cookie"})
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
