package admin

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	logger := logging.NewNoopLogger()

	rpm := mockrouting.NewRouteParamManager()
	rpm.On(
		"BuildRouteParamStringIDFetcher",
		UserIDURIParamKey,
	).Return(func(*http.Request) string { return "" })

	s := ProvideService(
		logger,
		&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
		&authentication.MockAuthenticator{},
		&mocktypes.AdminUserDataManager{},
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

		logger := logging.NewNoopLogger()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamStringIDFetcher",
			UserIDURIParamKey,
		).Return(func(*http.Request) string { return "" })

		s := ProvideService(
			logger,
			&authservice.Config{Cookies: authservice.CookieConfig{SigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"}},
			&authentication.MockAuthenticator{},
			&mocktypes.AdminUserDataManager{},
			scs.New(),
			encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON),
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
