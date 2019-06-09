package auth

import (
	"net/http"
	"testing"

	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	logger := noop.ProvideNoopLogger()
	cfg := &config.ServerConfig{
		Auth: config.AuthSettings{
			CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
		},
	}
	auth := &mauth.Authenticator{}
	userDB := &mmodels.UserDataManager{}
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
