package frontend

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

func TestProvideFrontendService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ProvideService(logging.NewNonOperationalLogger(), Config{})
	})
}
