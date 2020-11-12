package httpserver

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestServer() *Server {
	s := &Server{
		DebugMode:  true,
		db:         database.BuildMockDatabase(),
		config:     &config.ServerConfig{},
		encoder:    &mockencoding.EncoderDecoder{},
		httpServer: provideHTTPServer(),
		logger:     noop.NewLogger(),
		frontendService: frontendservice.ProvideFrontendService(
			noop.NewLogger(),
			config.FrontendSettings{},
		),
		webhooksService:      &mockmodels.WebhookDataServer{},
		usersService:         &mockmodels.UserDataServer{},
		authService:          &authservice.Service{},
		itemsService:         &mockmodels.ItemDataServer{},
		oauth2ClientsService: &mockmodels.OAuth2ClientDataServer{},
	}

	return s
}

func TestProvideServer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		actual, err := ProvideServer(
			ctx,
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISISAVERYLONGSTRINGFORTESTPURPOSES",
				},
			},
			&authservice.Service{},
			&frontendservice.Service{},
			&auditservice.Service{},
			&itemsservice.Service{},
			&usersservice.Service{},
			&oauth2clientsservice.Service{},
			&webhooksservice.Service{},
			database.BuildMockDatabase(),
			noop.NewLogger(),
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid cookie secret", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		actual, err := ProvideServer(
			ctx,
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISSTRINGISNTLONGENOUGH:(",
				},
			},
			&authservice.Service{},
			&frontendservice.Service{},
			&auditservice.Service{},
			&itemsservice.Service{},
			&usersservice.Service{},
			&oauth2clientsservice.Service{},
			&webhooksservice.Service{},
			database.BuildMockDatabase(),
			noop.NewLogger(),
			&mockencoding.EncoderDecoder{},
		)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
