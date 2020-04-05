package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
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
		return fakemodels.BuildFakeUser().ID
	}
	ed := encoding.ProvideResponseEncoder()

	service, err := ProvideAuthService(
		logger,
		cfg,
		auth,
		userDB,
		oauth,
		userIDFetcher,
		ed,
	)
	require.NoError(t, err)

	return service
}

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		cfg := &config.ServerConfig{
			Auth: config.AuthSettings{
				CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
			},
		}
		auth := &mockauth.Authenticator{}
		userDB := &mockmodels.UserDataManager{}
		oauth := &mockOAuth2ClientValidator{}
		userIDFetcher := func(*http.Request) uint64 {
			return fakemodels.BuildFakeUser().ID
		}
		ed := encoding.ProvideResponseEncoder()

		service, err := ProvideAuthService(
			noop.ProvideNoopLogger(),
			cfg,
			auth,
			userDB,
			oauth,
			userIDFetcher,
			ed,
		)
		assert.NotNil(t, service)
		assert.NoError(t, err)
	})

	T.Run("happy path", func(t *testing.T) {
		auth := &mockauth.Authenticator{}
		userDB := &mockmodels.UserDataManager{}
		oauth := &mockOAuth2ClientValidator{}
		userIDFetcher := func(*http.Request) uint64 {
			return fakemodels.BuildFakeUser().ID
		}
		ed := encoding.ProvideResponseEncoder()

		service, err := ProvideAuthService(
			noop.ProvideNoopLogger(),
			nil,
			auth,
			userDB,
			oauth,
			userIDFetcher,
			ed,
		)
		assert.Nil(t, service)
		assert.Error(t, err)
	})
}
