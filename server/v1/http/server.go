package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
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

		// Auth stuff
		adminUserExists bool
	}
)

// ProvideServer builds a new Server instance
func ProvideServer(
	config *config.ServerConfig,

	// services
	authService *auth.Service,
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

	if len(config.Auth.CookieSecret) < 32 {
		err := errors.New("cookie secret is too short, must be at least 32 characters in length")
		logger.Error(err, "cookie secret failure")
		return nil, err
	}

	srv := &Server{
		DebugMode: config.Server.Debug,

		// infra things
		db:          db,
		config:      config.Server,
		encoder:     encoder,
		httpServer:  provideHTTPServer(),
		logger:      logger.WithName("api_server"),
		newsManager: newsManager,

		// services
		webhooksService:      webhooksService,
		usersService:         usersService,
		authService:          authService,
		itemsService:         itemsService,
		oauth2ClientsService: oauth2Service,
	}

	ih, err := config.ProvideInstrumentationHandler(logger)
	if err != nil {
		return nil, err
	}

	if err := config.ProvideTracing(logger); err != nil {
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

	srv.setupRouter(config.Server.FrontendFilesDirectory, ih)
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
	log.Fatal(s.httpServer.ListenAndServe())
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

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "Unexpected internal error occurred"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
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

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "Unauthorized"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
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

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "invalid input"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
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

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "not found"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
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

	if err = s.encoder.EncodeResponse(res, genericResponse{Status: sc, Message: "invalid cookie"}); err != nil {
		s.logger.Error(err, "encoding response")
	}
}
