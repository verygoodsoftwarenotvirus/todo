package httpserver

import (
	"testing"

	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func TestProvideServer(T *testing.T) {
	T.SkipNow()

	T.Run("happy path", func(t *testing.T) {
		t.SkipNow()

		actual, err := ProvideServer(
			Config{},
			frontendservice.Config{},
			metrics.Config{},
			nil,
			&mocktypes.AuthService{},
			&mocktypes.FrontendService{},
			&mocktypes.AuditLogDataService{},
			&mocktypes.ItemDataServer{},
			&mocktypes.UserDataServer{},
			&mocktypes.PlanDataServer{},
			&mocktypes.APIClientDataServer{},
			&mocktypes.WebhookDataServer{},
			&mocktypes.AdminServer{},
			database.BuildMockDatabase(),
			logging.NewNonOperationalLogger(),
			mockencoding.NewMockEncoderDecoder(),
			chi.NewRouter(logging.NewNonOperationalLogger()),
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
