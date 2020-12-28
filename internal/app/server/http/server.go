package httpserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-chi/chi"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	serverNamespace = "todo_service"
	loggerName      = "api_server"
)

type (
	// Server is our API httpServer.
	Server struct {
		// Services.
		authService          types.AuthService
		frontendService      types.FrontendService
		auditService         types.AuditLogDataService
		usersService         types.UserDataService
		adminService         types.AdminService
		oauth2ClientsService types.OAuth2ClientDataService
		webhooksService      types.WebhookDataService
		itemsService         types.ItemDataService

		// infra things.
		db               database.DataManager
		serverSettings   Config
		frontendSettings frontendservice.Config
		router           *chi.Mux
		httpServer       *http.Server
		logger           logging.Logger
		encoder          encoding.EncoderDecoder
		tracer           tracing.Tracer
	}
)

// ProvideServer builds a new Server instance.
func ProvideServer(
	serverSettings Config,
	frontendSettings frontendservice.Config,
	metricsHandler metrics.InstrumentationHandler,
	authService types.AuthService,
	frontendService types.FrontendService,
	auditService types.AuditLogDataService,
	itemsService types.ItemDataService,
	usersService types.UserDataService,
	oauth2Service types.OAuth2ClientDataService,
	webhooksService types.WebhookDataService,
	adminService types.AdminService,
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
		httpServer:       provideHTTPServer(serverSettings.HTTPPort),
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
		tracer:               tracing.NewTracer(loggerName),
	}

	srv.setupRouter(metricsHandler)

	logger.Debug("HTTP server successfully constructed")

	return srv, nil
}

// Serve serves HTTP traffic.
func (s *Server) Serve() {
	s.logger.Debug("setting up opentelemetry handler")

	s.httpServer.Handler = otelhttp.NewHandler(
		s.router,
		serverNamespace,
		otelhttp.WithSpanNameFormatter(formatSpanNameForRequest),
	)

	s.logger.WithValue("listening_on", s.httpServer.Addr).Debug("Listening for HTTP requests")

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
func provideHTTPServer(port uint16) *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
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
