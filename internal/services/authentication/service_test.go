package authentication

import (
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockauthn "gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNoopLogger()
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
		&mockauthn.Authenticator{},
		&mocktypes.UserDataManager{},
		&mocktypes.APIClientDataManager{},
		&mocktypes.AccountUserMembershipDataManager{},
		scs.New(),
		encoderDecoder,
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNoopLogger()
		encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		s, err := ProvideService(
			logger,
			&Config{
				Cookies: CookieConfig{
					Name:       DefaultCookieName,
					SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
				},
			},
			&mockauthn.Authenticator{},
			&mocktypes.UserDataManager{},
			&mocktypes.APIClientDataManager{},
			&mocktypes.AccountUserMembershipDataManager{},
			scs.New(),
			encoderDecoder,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with invalid cookie key", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNoopLogger()
		encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		s, err := ProvideService(
			logger,
			&Config{
				Cookies: CookieConfig{
					Name:       DefaultCookieName,
					SigningKey: "BLAHBLAHBLAH",
				},
			},
			&mockauthn.Authenticator{},
			&mocktypes.UserDataManager{},
			&mocktypes.APIClientDataManager{},
			&mocktypes.AccountUserMembershipDataManager{},
			scs.New(),
			encoderDecoder,
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
