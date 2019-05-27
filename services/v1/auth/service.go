package auth

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"

	"github.com/gorilla/securecookie"
	libauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	serviceName = "auth_service"
)

type (
	oauth2ClientValidator interface {
		RequestIsAuthenticated(req *http.Request) (*models.OAuth2Client, error)
	}

	// Service handles auth
	Service struct {
		authenticator        libauth.Authenticator
		logger               logging.Logger
		userIDFetcher        UserIDFetcher
		database             models.UserDataManager
		oauth2ClientsService oauth2ClientValidator
		encoder              encoding.EncoderDecoder
		cookieBuilder        *securecookie.SecureCookie
	}
)

// UserIDFetcher is a function that fetches user IDs
type UserIDFetcher func(*http.Request) uint64

// ProvideAuthService builds a new AuthService
func ProvideAuthService(
	logger logging.Logger,
	config *config.ServerConfig,
	authenticator libauth.Authenticator,
	database models.UserDataManager,
	oauth2ClientsService oauth2ClientValidator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
) *Service {
	svc := &Service{
		logger:               logger.WithName(serviceName),
		encoder:              encoder,
		database:             database,
		oauth2ClientsService: oauth2ClientsService,
		authenticator:        authenticator,
		userIDFetcher:        userIDFetcher,
		cookieBuilder: securecookie.New(
			securecookie.GenerateRandomKey(64),
			[]byte(config.Auth.CookieSecret),
		),
	}

	return svc
}
