package auth

import (
	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

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
		config.AuthSettings{CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
		&mockauth.Authenticator{},
		&mockmodels.UserDataManager{},
		&mockmodels.AuditLogEntryDataManager{},
		&mockOAuth2ClientValidator{},
		scs.New(),
		ed,
		func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
	)
	require.NoError(t, err)

	return service
}

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		logger := noop.NewLogger()
		ed := encoding.ProvideResponseEncoder(logger)

		service, err := ProvideAuthService(
			logger,
			config.AuthSettings{CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!"},
			&mockauth.Authenticator{},
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogEntryDataManager{},
			&mockOAuth2ClientValidator{},
			scs.New(),
			ed,
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
		)
		assert.NotNil(t, service)
		assert.NoError(t, err)
	})
}
