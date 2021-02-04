package admin

import (
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/routeparams"
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
		&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
		&mockauth.Authenticator{},
		&mocktypes.AdminUserDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		scs.New(),
		ed,
		routeparams.NewRouteParamManager(),
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideAdminService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		logger := logging.NewNonOperationalLogger()
		ed := encoding.ProvideEncoderDecoder(logger)

		s, err := ProvideService(
			logger,
			&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
			&mockauth.Authenticator{},
			&mocktypes.AdminUserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			scs.New(),
			ed,
			routeparams.NewRouteParamManager(),
		)
		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}
