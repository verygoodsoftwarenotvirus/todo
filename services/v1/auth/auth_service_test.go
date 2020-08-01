package auth

import (
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	logger := noop.ProvideNoopLogger()
	cfg := config.AuthSettings{
		CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
	}
	auth := &mockauth.Authenticator{}
	userDB := &mockmodels.UserDataManager{}
	oauth := &mockOAuth2ClientValidator{}
	ed := encoding.ProvideResponseEncoder()

	sm := scs.New()
	// this is currently the default, but in case that changes
	sm.Store = memstore.New()

	service, err := ProvideAuthService(
		logger,
		cfg,
		auth,
		userDB,
		oauth,
		sm,
		ed,
	)
	require.NoError(t, err)

	return service
}

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		cfg := config.AuthSettings{
			CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
		}
		auth := &mockauth.Authenticator{}
		userDB := &mockmodels.UserDataManager{}
		oauth := &mockOAuth2ClientValidator{}
		ed := encoding.ProvideResponseEncoder()
		sm := scs.New()

		service, err := ProvideAuthService(
			noop.ProvideNoopLogger(),
			cfg,
			auth,
			userDB,
			oauth,
			sm,
			ed,
		)
		assert.NotNil(t, service)
		assert.NoError(t, err)
	})
}
