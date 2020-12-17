package auth

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/securecookie"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	// CookieName is the name of the cookie we attach to requests.
	CookieName         = "todocookie"
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

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// service handles authentication service-wide.
	service struct {
		config               config.AuthSettings
		logger               logging.Logger
		authenticator        password.Authenticator
		userDB               types.UserDataManager
		auditLog             types.AuthAuditManager
		oauth2ClientsService types.OAuth2ClientDataService
		encoderDecoder       encoding.EncoderDecoder
		cookieManager        cookieEncoderDecoder
		sessionManager       *scs.SessionManager
		sessionInfoFetcher   SessionInfoFetcher
	}
)

// ProvideService builds a new AuthService.
func ProvideService(
	logger logging.Logger,
	cfg config.AuthSettings,
	authenticator password.Authenticator,
	userDataManager types.UserDataManager,
	auditLog types.AuthAuditManager,
	oauth2ClientsService types.OAuth2ClientDataService,
	sessionManager *scs.SessionManager,
	encoder encoding.EncoderDecoder,
	sessionInfoFetcher SessionInfoFetcher,
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
		sessionInfoFetcher:   sessionInfoFetcher,
		cookieManager: securecookie.New(
			securecookie.GenerateRandomKey(cookieSecretSize),
			[]byte(cfg.CookieSigningKey),
		),
	}
	svc.sessionManager.Lifetime = cfg.CookieLifetime

	if _, err := svc.cookieManager.Encode(CookieName, "blah"); err != nil {
		logger.WithValue("cookie_signing_key", cfg.CookieSigningKey).Error(err, "building test cookie")
		return nil, fmt.Errorf("error building test cookie: %w", err)
	}

	return svc, nil
}
