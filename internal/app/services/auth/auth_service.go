package auth

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
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
	// OAuth2ClientValidator is a stand-in interface, where we needed to abstract
	// a regular structure with an interface for testing purposes.
	OAuth2ClientValidator interface {
		ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*types.OAuth2Client, error)
	}

	// cookieEncoderDecoder is a stand-in interface for gorilla/securecookie.
	cookieEncoderDecoder interface {
		Encode(name string, value interface{}) (string, error)
		Decode(name, value string, dst interface{}) error
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// Service handles authentication service-wide.
	Service struct {
		config               config.AuthSettings
		logger               logging.Logger
		authenticator        auth.Authenticator
		userDB               types.UserDataManager
		auditLog             types.AuthAuditManager
		oauth2ClientsService OAuth2ClientValidator
		encoderDecoder       encoding.EncoderDecoder
		cookieManager        cookieEncoderDecoder
		sessionManager       *scs.SessionManager
		sessionInfoFetcher   SessionInfoFetcher
	}
)

// ProvideAuthService builds a new AuthService.
func ProvideAuthService(
	logger logging.Logger,
	cfg config.AuthSettings,
	authenticator auth.Authenticator,
	userDataManager types.UserDataManager,
	auditLog types.AuthAuditManager,
	oauth2ClientsService OAuth2ClientValidator,
	sessionManager *scs.SessionManager,
	encoder encoding.EncoderDecoder,
	sessionInfoFetcher SessionInfoFetcher,
) (*Service, error) {
	svc := &Service{
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
