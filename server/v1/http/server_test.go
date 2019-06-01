package httpserver

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/stretchr/testify/assert"
)

func buildTestServer() *Server {
	s := &Server{
		DebugMode:  true,
		db:         database.BuildMockDatabase(),
		config:     config.ServerSettings{},
		encoder:    &mencoding.EncoderDecoder{},
		httpServer: provideHTTPServer(),
		logger:     noop.ProvideNoopLogger(),

		frontendService:      frontend.ProvideFrontendService(noop.ProvideNoopLogger()),
		webhooksService:      &mmodels.WebhookDataServer{},
		usersService:         &mmodels.UserDataServer{},
		authService:          &auth.Service{},
		itemsService:         &mmodels.ItemDataServer{},
		oauth2ClientsService: &mmodels.OAuth2ClientDataServer{},
	}
	return s
}

func TestProvideServer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()

		actual, err := ProvideServer(
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
			&mencoding.EncoderDecoder{},
			newsman.NewNewsman(nil, nil),
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

}
