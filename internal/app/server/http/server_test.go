package httpserver

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestServer() *Server {
	s := &Server{
		db:               database.BuildMockDatabase(),
		serverSettings:   config.ServerSettings{},
		frontendSettings: config.FrontendSettings{},
		encoder:          &mockencoding.EncoderDecoder{},
		httpServer:       provideHTTPServer(),
		logger:           noop.NewLogger(),
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

		actual, err := ProvideServer(
			config.ServerSettings{},
			config.FrontendSettings{},
			nil,
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
}
