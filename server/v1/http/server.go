package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// Server is our API httpServer
	Server struct {
		DebugMode bool

		// Services
		authService          *auth.Service
		frontendService      *frontend.Service
		itemsService         models.ItemDataServer
		usersService         models.UserDataServer
		oauth2ClientsService models.Oauth2ClientDataServer
		webhooksService      models.WebhookDataServer

		// infra things
		db          database.Database
		config      config.ServerSettings
		router      *chi.Mux
		httpServer  *http.Server
		logger      logging.Logger
		encoder     encoding.EncoderDecoder
		newsManager *newsman.Newsman
	}
)

// ProvideServer builds a new Server instance
func ProvideServer(
	cfg *config.ServerConfig,

	// services
	authService *auth.Service,
	frontendService *frontend.Service,
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,
	webhooksService *webhooks.Service,

	// infra things
	db database.Database,
	logger logging.Logger,
	encoder encoding.EncoderDecoder,
	newsManager *newsman.Newsman,
) (*Server, error) {
	ctx := context.Background()

	if len(cfg.Auth.CookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	srv := &Server{
		DebugMode: cfg.Server.Debug,

		// infra things
		db:          db,
		config:      cfg.Server,
		encoder:     encoder,
		httpServer:  provideHTTPServer(),
		logger:      logger.WithName("api_server"),
		newsManager: newsManager,

		// services
		webhooksService:      webhooksService,
		frontendService:      frontendService,
		usersService:         usersService,
		authService:          authService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	ih, err := cfg.ProvideInstrumentationHandler(logger)
	if err != nil && err != config.ErrInvalidMetricsProvider {
		return nil, err
	}

	if err = cfg.ProvideTracing(logger); err != nil && err != config.ErrInvalidTracingProvider {
		return nil, err
	}

	allWebhooks, err := db.GetAllWebhooks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "initializing webhooks")
	}
	for _, wh := range allWebhooks.Webhooks {
		// NOTE: we must guarantee that whatever is stored in the database is valid, otherwise
		// newsman will try (and fail) to execute requests constantly
		l := wh.ToListener(srv.logger)
		srv.newsManager.TuneIn(l)
	}

	srv.setupRouter(cfg.Frontend, ih)
	srv.httpServer.Handler = &ochttp.Handler{
		Handler:        srv.router,
		FormatSpanName: formatSpanNameForRequest,
	}

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.httpServer.Addr = fmt.Sprintf(":%d", s.config.HTTPPort)
	s.logger.Debug(fmt.Sprintf("Listening for HTTP requests on %s", s.httpServer.Addr))

	// returns ErrServerClosed on graceful close
	if err := s.httpServer.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop.
			os.Exit(0)
		}
	}
}

// ErrorNotifier is a function which can notify a user of an error
type ErrorNotifier func(res http.ResponseWriter, req *http.Request, err error)
