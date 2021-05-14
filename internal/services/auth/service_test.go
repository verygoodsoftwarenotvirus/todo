package auth

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNonOperationalLogger()
	encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

	s, err := ProvideService(
		logger,
		&Config{
			Cookies: CookieConfig{
				Name:       DefaultCookieName,
				SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
			},
			PASETO: PASETOConfig{
				Issuer:       "test",
				LocalModeKey: []byte("BLAHBLAHBLAHPRETENDTHISISSECRET!"),
				Lifetime:     time.Hour,
			},
		},
		&passwords.MockAuthenticator{},
		&mocktypes.UserDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		&mocktypes.APIClientDataManager{},
		&mocktypes.AccountUserMembershipDataManager{},
		scs.New(),
		encoderDecoder,
		chi.NewRouteParamManager(),
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNonOperationalLogger()
		encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		s, err := ProvideService(
			logger,
			&Config{
				Cookies: CookieConfig{
					Name:       DefaultCookieName,
					SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
				},
			},
			&passwords.MockAuthenticator{},
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mocktypes.APIClientDataManager{},
			&mocktypes.AccountUserMembershipDataManager{},
			scs.New(),
			encoderDecoder,
			chi.NewRouteParamManager(),
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with invalid cookie key", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNonOperationalLogger()
		encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		s, err := ProvideService(
			logger,
			&Config{
				Cookies: CookieConfig{
					Name:       DefaultCookieName,
					SigningKey: "BLAHBLAHBLAH",
				},
			},
			&passwords.MockAuthenticator{},
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mocktypes.APIClientDataManager{},
			&mocktypes.AccountUserMembershipDataManager{},
			scs.New(),
			encoderDecoder,
			chi.NewRouteParamManager(),
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
