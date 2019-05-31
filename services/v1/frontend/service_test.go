package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"testing"
)

func TestProvideFrontendService(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideFrontendService(noop.ProvideNoopLogger())
	})
}
