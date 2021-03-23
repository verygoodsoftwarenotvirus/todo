package httpserver

import (
	"context"
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

	T.Run("standard", func(t *testing.T) {
		t.SkipNow()

		actual, err := ProvideServer(
			context.Background(),
			Config{},
			frontendservice.Config{},
			metrics.Config{},
			nil,
			&mocktypes.AuthService{},
			&mocktypes.AuditLogDataService{},
			&mocktypes.UserDataServer{},
			&mocktypes.AccountDataServer{},
			&mocktypes.PlanDataServer{},
			&mocktypes.APIClientDataServer{},
			&mocktypes.ItemDataServer{},
			&mocktypes.WebhookDataServer{},
			&mocktypes.AdminServer{},
			&mocktypes.FrontendService{},
			database.BuildMockDatabase(),
			logging.NewNonOperationalLogger(),
			mockencoding.NewMockEncoderDecoder(),
			chi.NewRouter(logging.NewNonOperationalLogger()),
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
