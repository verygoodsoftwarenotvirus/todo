package auth

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	logger := noop.NewLogger()
	ed := encoding.ProvideResponseEncoder(logger)

	service, err := ProvideAuthService(
		logger,
		config.AuthSettings{CookieSigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
		&mockauth.Authenticator{},
		&mockmodels.UserDataManager{},
		&mockmodels.AuditLogDataManager{},
		&mockOAuth2ClientValidator{},
		scs.New(),
		ed,
		func(*http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
	)
	require.NoError(t, err)

	return service
}

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		logger := noop.NewLogger()
		ed := encoding.ProvideResponseEncoder(logger)

		service, err := ProvideAuthService(
			logger,
			config.AuthSettings{CookieSigningKey: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
			&mockauth.Authenticator{},
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockOAuth2ClientValidator{},
			scs.New(),
			ed,
			func(*http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
		)
		assert.NotNil(t, service)
		assert.NoError(t, err)
	})
}
