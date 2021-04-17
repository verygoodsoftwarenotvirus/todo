package auth

import (
	"testing"
	"time"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

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
		&mockauth.Authenticator{},
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
			&mockauth.Authenticator{},
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
			&mockauth.Authenticator{},
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
