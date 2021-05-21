package frontend

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/mock"
)

func TestService_SetupRoutes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		obligatoryHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

		authService := &mocktypes.AuthService{}
		authService.On(
			"ServiceAdminMiddleware",
			mock.IsType(obligatoryHandler),
		).Return(http.Handler(obligatoryHandler))

		authService.On(
			"UserAttributionMiddleware",
			mock.IsType(obligatoryHandler),
		).Return(http.Handler(obligatoryHandler))
		s.authService = authService

		router := chi.NewRouter(logging.NewNonOperationalLogger())

		s.SetupRoutes(router)
	})
}
