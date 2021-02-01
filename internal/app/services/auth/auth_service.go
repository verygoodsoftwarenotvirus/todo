package auth

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/securecookie"
)

const (
	serviceName        = "auth_service"
	sessionInfoKey     = string(types.SessionInfoKey)
	cookieErrorLogName = "_COOKIE_CONSTRUCTION_ERROR_"
	cookieSecretSize   = 64
)

type (
	// cookieEncoderDecoder is a stand-in interface for gorilla/securecookie.
	cookieEncoderDecoder interface {
		Encode(name string, value interface{}) (string, error)
		Decode(name, value string, dst interface{}) error
	}

	// service handles authentication service-wide.
	service struct {
		config               *Config
		logger               logging.Logger
		authenticator        password.Authenticator
		userDB               types.UserDataManager
		auditLog             types.AuthAuditManager
		oauth2ClientsService types.OAuth2ClientDataService
		encoderDecoder       encoding.EncoderDecoder
		cookieManager        cookieEncoderDecoder
		sessionManager       *scs.SessionManager
		sessionInfoFetcher   func(*http.Request) (*types.SessionInfo, error)
		tracer               tracing.Tracer
	}
)

// ProvideService builds a new AuthService.
func ProvideService(
	logger logging.Logger,
	cfg *Config,
	authenticator password.Authenticator,
	userDataManager types.UserDataManager,
	auditLog types.AuthAuditManager,
	oauth2ClientsService types.OAuth2ClientDataService,
	sessionManager *scs.SessionManager,
	encoder encoding.EncoderDecoder,
) (types.AuthService, error) {
	svc := &service{
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoder,
		config:               cfg,
		userDB:               userDataManager,
		auditLog:             auditLog,
		oauth2ClientsService: oauth2ClientsService,
		authenticator:        authenticator,
		sessionManager:       sessionManager,
		sessionInfoFetcher:   routeparams.SessionInfoFetcherFromRequestContext,
		cookieManager: securecookie.New(
			securecookie.GenerateRandomKey(cookieSecretSize),
			[]byte(cfg.Cookies.SigningKey),
		),
		tracer: tracing.NewTracer(serviceName),
	}
	svc.sessionManager.Lifetime = cfg.Cookies.Lifetime

	if _, err := svc.cookieManager.Encode(cfg.Cookies.Name, "blah"); err != nil {
		logger.WithValue("cookie_signing_key_length", len(cfg.Cookies.SigningKey)).Error(err, "building test cookie")
		return nil, fmt.Errorf("building test cookie: %w", err)
	}

	return svc, nil
}
