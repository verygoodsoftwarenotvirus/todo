package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"

	"github.com/go-chi/chi"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opencensus.io/plugin/ochttp"
)

const (
	maxTimeout      = 120 * time.Second
	serverNamespace = "todo-service"
	loggerName      = "api_server"
)

type (
	// Server is our API httpServer.
	Server struct {
		DebugMode bool

		// Services.
		authService          *authservice.Service
		frontendService      *frontendservice.Service
		auditService         models.AuditLogDataServer
		usersService         models.UserDataServer
		oauth2ClientsService models.OAuth2ClientDataServer
		webhooksService      models.WebhookDataServer
		itemsService         models.ItemDataServer

		// infra things.
		db         database.DataManager
		config     *config.ServerConfig
		router     *chi.Mux
		httpServer *http.Server
		logger     logging.Logger
		encoder    encoding.EncoderDecoder
	}
)

// ProvideServer builds a new Server instance.
func ProvideServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	authService *authservice.Service,
	frontendService *frontendservice.Service,
	auditService models.AuditLogDataServer,
	itemsService models.ItemDataServer,
	usersService models.UserDataServer,
	oauth2Service models.OAuth2ClientDataServer,
	webhooksService models.WebhookDataServer,
	db database.DataManager,
	logger logging.Logger,
	encoder encoding.EncoderDecoder,
) (*Server, error) {
	if len(cfg.Auth.CookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	srv := &Server{
		DebugMode: cfg.Server.Debug,
		// infra things,
		db:         db,
		config:     cfg,
		encoder:    encoder,
		httpServer: provideHTTPServer(),
		logger:     logger.WithName(loggerName),
		// services,
		auditService:         auditService,
		webhooksService:      webhooksService,
		frontendService:      frontendService,
		usersService:         usersService,
		authService:          authService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	if err := cfg.ProvideTracing(logger); err != nil && err != config.ErrInvalidTracingProvider {
		return nil, err
	}

	metricsHandler := cfg.ProvideInstrumentationHandler(logger)
	srv.setupRouter(cfg, metricsHandler)

	srv.httpServer.Handler = &ochttp.Handler{
		Handler:        srv.router,
		FormatSpanName: formatSpanNameForRequest,
	}

	return srv, nil
}

/*
func (s *Server) logRoutes() {
	if err := chi.Walk(s.router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		s.logger.WithValues(map[string]interface{}{
			"method": method,
			"route":  route,
		}).Debug("route found")

		return nil
	}); err != nil {
		s.logger.Error(err, "logging routes")
	}
}
*/

// Serve serves HTTP traffic.
func (s *Server) Serve() {
	s.httpServer.Addr = fmt.Sprintf(":%d", s.config.Server.HTTPPort)
	s.logger.Debug(fmt.Sprintf("Listening for HTTP requests on %q", s.httpServer.Addr))

	// returns ErrServerClosed on graceful close.
	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Error(err, "server shutting down")
		if err == http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop.
			os.Exit(0)
		}
	}
}

// provideHTTPServer provides an HTTP httpServer.
func provideHTTPServer() *http.Server {
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
