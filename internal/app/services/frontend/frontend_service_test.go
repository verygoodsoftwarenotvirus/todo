package frontend

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideFrontendService(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ProvideService(noop.NewLogger(), Config{})
	})
}
