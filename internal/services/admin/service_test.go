package admin

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNonOperationalLogger()

	rpm := mockrouting.NewRouteParamManager()
	rpm.On(
		"BuildRouteParamIDFetcher",
		mock.IsType(logging.NewNonOperationalLogger()), UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

	s := ProvideService(
		logger,
		&authentication.Config{Cookies: authentication.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
		&passwords.MockAuthenticator{},
		&mocktypes.AdminUserDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		scs.New(),
		encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON),
		rpm,
	)

	mock.AssertExpectationsForObjects(t, rpm)

	srv, ok := s.(*service)
	require.True(t, ok)

	return srv
}

func TestProvideAdminService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.IsType(logging.NewNonOperationalLogger()), UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideService(
			logger,
			&authentication.Config{Cookies: authentication.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
			&passwords.MockAuthenticator{},
			&mocktypes.AdminUserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			scs.New(),
			encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON),
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
