package auth

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/securecookie"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	serviceName = "auth_service"
)

type (
	// OAuth2ClientValidator is a stand-in interface, where we needed to abstract
	// a regular structure with an interface for testing purposes
	OAuth2ClientValidator interface {
		ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*models.OAuth2Client, error)
	}

	// cookieEncoderDecoder is a stand-in interface for gorilla/securecookie
	cookieEncoderDecoder interface {
		Encode(name string, value interface{}) (string, error)
		Decode(name, value string, dst interface{}) error
	}

	// UserIDFetcher is a function that fetches user IDs
	UserIDFetcher func(*http.Request) uint64

	// Service handles authentication service-wide
	Service struct {
		config               config.AuthSettings
		logger               logging.Logger
		authenticator        auth.Authenticator
		userIDFetcher        UserIDFetcher
		userDB               models.UserDataManager
		oauth2ClientsService OAuth2ClientValidator
		encoderDecoder       encoding.EncoderDecoder
		cookieManager        cookieEncoderDecoder
	}
)

// ProvideAuthService builds a new AuthService
func ProvideAuthService(
	logger logging.Logger,
	cfg *config.ServerConfig,
	authenticator auth.Authenticator,
	database models.UserDataManager,
	oauth2ClientsService OAuth2ClientValidator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
) *Service {
	svc := &Service{
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoder,
		config:               cfg.Auth,
		userDB:               database,
		oauth2ClientsService: oauth2ClientsService,
		authenticator:        authenticator,
		userIDFetcher:        userIDFetcher,
		cookieManager: securecookie.New(
			securecookie.GenerateRandomKey(64),
			[]byte(cfg.Auth.CookieSecret),
		),
	}

	return svc
}
