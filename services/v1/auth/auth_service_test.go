package auth

import (
	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	logger := noop.ProvideNoopLogger()
	cfg := &config.ServerConfig{
		Auth: config.AuthSettings{
			CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
		},
	}
	auth := &mockauth.Authenticator{}
	userDB := &mockmodels.UserDataManager{}
	oauth := &mockOAuth2ClientValidator{}
	userIDFetcher := func(*http.Request) uint64 {
		return 1
	}
	ed := encoding.ProvideResponseEncoder()

	service := ProvideAuthService(
		logger,
		cfg,
		auth,
		userDB,
		oauth,
		userIDFetcher,
		ed,
	)

	return service
}
