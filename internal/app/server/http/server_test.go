package httpserver

import (
	"testing"

	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestServer() *Server {
	l := noop.NewLogger()

	s := &Server{
		logger:               l,
		db:                   database.BuildMockDatabase(),
		serverSettings:       Config{},
		frontendSettings:     frontendservice.Config{},
		encoder:              &mockencoding.EncoderDecoder{},
		httpServer:           provideHTTPServer(),
		frontendService:      &mocktypes.FrontendService{},
		webhooksService:      &mocktypes.WebhookDataServer{},
		usersService:         &mocktypes.UserDataServer{},
		authService:          &mocktypes.AuthService{},
		itemsService:         &mocktypes.ItemDataServer{},
		oauth2ClientsService: &mocktypes.OAuth2ClientDataServer{},
	}

	return s
}

func TestProvideServer(T *testing.T) {
	T.SkipNow()

	T.Run("happy path", func(t *testing.T) {
		t.SkipNow()

		actual, err := ProvideServer(
			Config{},
			frontendservice.Config{},
			nil,
			&mocktypes.AuthService{},
			&mocktypes.FrontendService{},
			&mocktypes.AuditLogDataService{},
			&mocktypes.ItemDataServer{},
			&mocktypes.UserDataServer{},
			&mocktypes.OAuth2ClientDataServer{},
			&mocktypes.WebhookDataServer{},
			&mocktypes.AdminServer{},
			database.BuildMockDatabase(),
			noop.NewLogger(),
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
