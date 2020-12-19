package admin

import (
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := noop.NewLogger()
	ed := encoding.ProvideResponseEncoder(logger)

	s, err := ProvideService(
		logger,
		authservice.Config{CookieSigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
		&mockauth.Authenticator{},
		&mocktypes.AdminUserDataManager{},
		&mocktypes.AuditLogDataManager{},
		scs.New(),
		ed,
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideAdminService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		logger := noop.NewLogger()
		ed := encoding.ProvideResponseEncoder(logger)

		service, err := ProvideService(
			logger,
			authservice.Config{CookieSigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
			&mockauth.Authenticator{},
			&mocktypes.AdminUserDataManager{},
			&mocktypes.AuditLogDataManager{},
			scs.New(),
			ed,
		)
		assert.NotNil(t, service)
		assert.NoError(t, err)
	})
}
