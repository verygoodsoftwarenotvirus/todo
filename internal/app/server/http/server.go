package httpserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-chi/chi"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serverNamespace = "todo-service"
	loggerName      = "api_server"
)

type (
	// Server is our API httpServer.
	Server struct {
		// Services.
		authService          *authservice.Service
		frontendService      *frontendservice.Service
		auditService         types.AuditLogDataServer
		usersService         types.UserDataServer
		adminService         types.AdminServer
		oauth2ClientsService types.OAuth2ClientDataServer
		webhooksService      types.WebhookDataServer
		itemsService         types.ItemDataServer

		// infra things.
		db               database.DataManager
		serverSettings   config.ServerSettings
		frontendSettings config.FrontendSettings
		router           *chi.Mux
		httpServer       *http.Server
		logger           logging.Logger
		encoder          encoding.EncoderDecoder
	}
)

// ProvideServer builds a new Server instance.
func ProvideServer(
	serverSettings config.ServerSettings,
	frontendSettings config.FrontendSettings,
	metricsHandler metrics.InstrumentationHandler,
	authService *authservice.Service,
	frontendService *frontendservice.Service,
	auditService types.AuditLogDataServer,
	itemsService types.ItemDataServer,
	usersService types.UserDataServer,
	oauth2Service types.OAuth2ClientDataServer,
	webhooksService types.WebhookDataServer,
	adminService types.AdminServer,
	db database.DataManager,
	logger logging.Logger,
	encoder encoding.EncoderDecoder,
) (*Server, error) {
	srv := &Server{
		// infra things,
		db:               db,
		serverSettings:   serverSettings,
		frontendSettings: frontendSettings,
		encoder:          encoder,
		httpServer:       provideHTTPServer(),
		logger:           logger.WithName(loggerName),
		// services,
		adminService:         adminService,
		auditService:         auditService,
		webhooksService:      webhooksService,
		frontendService:      frontendService,
		usersService:         usersService,
		authService:          authService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	srv.setupRouter(metricsHandler)

	logger.Debug("HTTP server successfully constructed")

	return srv, nil
}

// Serve serves HTTP traffic.
func (s *Server) Serve() {
	s.httpServer.Addr = fmt.Sprintf(":%d", s.serverSettings.HTTPPort)
	s.logger.Debug(fmt.Sprintf("Listening for HTTP requests on %q", s.httpServer.Addr))

	// returns ErrServerClosed on graceful close.
	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Error(err, "server shutting down")

		if errors.Is(err, http.ErrServerClosed) {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop.
			os.Exit(0)
		}
	}
}

const (
	maxTimeout   = 120 * time.Second
	readTimeout  = 5 * time.Second
	writeTimeout = 2 * readTimeout
	idleTimeout  = maxTimeout
)

// provideHTTPServer provides an HTTP httpServer.
func provideHTTPServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	srv := &http.Server{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
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
