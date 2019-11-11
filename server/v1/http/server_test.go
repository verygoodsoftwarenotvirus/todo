package httpserver

import (
	"context"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

func buildTestServer() *Server {
	s := &Server{
		DebugMode:  true,
		db:         database.BuildMockDatabase(),
		config:     &config.ServerConfig{},
		encoder:    &mockencoding.EncoderDecoder{},
		httpServer: provideHTTPServer(),
		logger:     noop.ProvideNoopLogger(),
		frontendService: frontend.ProvideFrontendService(
			noop.ProvideNoopLogger(),
			config.FrontendSettings{},
		),
		webhooksService:      &mockmodels.WebhookDataServer{},
		usersService:         &mockmodels.UserDataServer{},
		authService:          &auth.Service{},
		itemsService:         &mockmodels.ItemDataServer{},
		oauth2ClientsService: &mockmodels.OAuth2ClientDataServer{},
	}
	return s
}

func TestProvideServer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(&models.WebhookList{}, nil)

		actual, err := ProvideServer(
			context.Background(),
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISISAVERYLONGSTRINGFORTESTPURPOSES",
				},
			},
			&auth.Service{},
			&frontend.Service{},
			&items.Service{},
			&users.Service{},
			&oauth2clients.Service{},
			&webhooks.Service{},
			mockDB,
			noop.ProvideNoopLogger(),
			&mockencoding.EncoderDecoder{},
			newsman.NewNewsman(nil, nil),
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
