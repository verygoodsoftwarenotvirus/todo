package frontend

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
)

func TestProvideFrontendService(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ProvideService(logging.NewNonOperationalLogger(), Config{})
	})
}
