package httpserver

import (
	"context"
	"errors"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

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
		frontendService: frontendservice.ProvideFrontendService(
			noop.ProvideNoopLogger(),
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
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(exampleWebhookList, nil)

		actual, err := ProvideServer(
			ctx,
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISISAVERYLONGSTRINGFORTESTPURPOSES",
				},
			},
			&authservice.Service{},
			&frontendservice.Service{},
			&itemsservice.Service{},
			&usersservice.Service{},
			&oauth2clientsservice.Service{},
			&webhooksservice.Service{},
			mockDB,
			noop.ProvideNoopLogger(),
			&mockencoding.EncoderDecoder{},
			newsman.NewNewsman(nil, nil),
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid cookie secret", func(t *testing.T) {
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(exampleWebhookList, nil)

		actual, err := ProvideServer(
			ctx,
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISSTRINGISNTLONGENOUGH:(",
				},
			},
			&authservice.Service{},
			&frontendservice.Service{},
			&itemsservice.Service{},
			&usersservice.Service{},
			&oauth2clientsservice.Service{},
			&webhooksservice.Service{},
			mockDB,
			noop.ProvideNoopLogger(),
			&mockencoding.EncoderDecoder{},
			newsman.NewNewsman(nil, nil),
		)

		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching webhooks", func(t *testing.T) {
		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return((*models.WebhookList)(nil), errors.New("blah"))

		actual, err := ProvideServer(
			ctx,
			&config.ServerConfig{
				Auth: config.AuthSettings{
					CookieSecret: "THISISAVERYLONGSTRINGFORTESTPURPOSES",
				},
			},
			&authservice.Service{},
			&frontendservice.Service{},
			&itemsservice.Service{},
			&usersservice.Service{},
			&oauth2clientsservice.Service{},
			&webhooksservice.Service{},
			mockDB,
			noop.ProvideNoopLogger(),
			&mockencoding.EncoderDecoder{},
			newsman.NewNewsman(nil, nil),
		)

		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
