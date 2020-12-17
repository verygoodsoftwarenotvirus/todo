package httpserver

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestServer() *Server {
	l := noop.NewLogger()

	s := &Server{
		logger:               l,
		db:                   database.BuildMockDatabase(),
		serverSettings:       config.ServerSettings{},
		frontendSettings:     config.FrontendSettings{},
		encoder:              &mockencoding.EncoderDecoder{},
		httpServer:           provideHTTPServer(),
		frontendService:      &mockmodels.FrontendService{},
		webhooksService:      &mockmodels.WebhookDataServer{},
		usersService:         &mockmodels.UserDataServer{},
		authService:          &mockmodels.AuthService{},
		itemsService:         &mockmodels.ItemDataServer{},
		oauth2ClientsService: &mockmodels.OAuth2ClientDataServer{},
	}

	return s
}

func TestProvideServer(T *testing.T) {
	T.SkipNow()

	T.Run("happy path", func(t *testing.T) {
		t.SkipNow()

		actual, err := ProvideServer(
			config.ServerSettings{},
			config.FrontendSettings{},
			nil,
			&mockmodels.AuthService{},
			&mockmodels.FrontendService{},
			&mockmodels.AuditLogDataService{},
			&mockmodels.ItemDataServer{},
			&mockmodels.UserDataServer{},
			&mockmodels.OAuth2ClientDataServer{},
			&mockmodels.WebhookDataServer{},
			&mockmodels.AdminServer{},
			database.BuildMockDatabase(),
			noop.NewLogger(),
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
