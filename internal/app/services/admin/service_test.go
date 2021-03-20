package admin

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNonOperationalLogger()

	rpm := mockrouting.NewRouteParamManager()
	rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

	s := ProvideService(
		logger,
		&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
		&mockauth.Authenticator{},
		&mocktypes.AdminUserDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		scs.New(),
		encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON),
		rpm,
	)

	mock.AssertExpectationsForObjects(t, rpm)

	return s.(*service)
}

func TestProvideAdminService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideService(
			logger,
			&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
			&mockauth.Authenticator{},
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
