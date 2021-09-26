package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func TestEnsureUnitCounter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ucp := func(string, string) UnitCounter {
			return &noopUnitCounter{}
		}

		assert.NotNil(t, EnsureUnitCounter(ucp, logging.NewNoopLogger(), "", ""))
	})

	T.Run("with nil UnitCounterProvider", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, EnsureUnitCounter(nil, logging.NewNoopLogger(), "", ""))
	})
}
