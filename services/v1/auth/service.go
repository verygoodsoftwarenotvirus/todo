package auth

import (
	"net/http"

	libauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/securecookie"
)

const (
	serviceName = "auth_service"
)

type (
	// OAuth2ClientValidator is a stand-in interface, where we needed to abstract
	// a regular structure with an interface for testing purposes
	OAuth2ClientValidator interface {
		RequestIsAuthenticated(req *http.Request) (*models.OAuth2Client, error)
	}

	cookieEncoderDecoder interface {
		Encode(name string, value interface{}) (string, error)
		Decode(name, value string, dst interface{}) error
	}

	// sessionManager interface {
	// // Set establishes the session
	// Set(context.Context, http.ResponseWriter, []byte, ...[]byte) error
	// // Clear clears the session
	// Clear(context.Context, http.ResponseWriter)
	// }

	// Service handles auth
	Service struct {
		logger               logging.Logger
		authenticator        libauth.Authenticator
		userIDFetcher        UserIDFetcher
		userDB               models.UserDataManager
		oauth2ClientsService OAuth2ClientValidator
		encoderDecoder       encoding.EncoderDecoder
		cookieBuilder        cookieEncoderDecoder
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
	oauth2ClientsService OAuth2ClientValidator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
) *Service {
	svc := &Service{
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoder,
		userDB:               database,
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
