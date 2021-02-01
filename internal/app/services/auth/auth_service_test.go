package auth

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNonOperationalLogger()
	ed := encoding.ProvideEncoderDecoder(logger)

	s, err := ProvideService(
		logger,
		&Config{
			Cookies: CookieConfig{
				Name:       DefaultCookieName,
				SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
			},
		},
		&mockauth.Authenticator{},
		&mocktypes.UserDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		&mocktypes.OAuth2ClientDataServer{},
		scs.New(),
		ed,
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNonOperationalLogger()
		ed := encoding.ProvideEncoderDecoder(logger)

		s, err := ProvideService(
			logger,
			&Config{
				Cookies: CookieConfig{
					Name:       DefaultCookieName,
					SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
				},
			},
			&mockauth.Authenticator{},
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mocktypes.OAuth2ClientDataServer{},
			scs.New(),
			ed,
		)
		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}
