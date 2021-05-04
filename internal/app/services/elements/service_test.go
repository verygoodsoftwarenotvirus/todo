package elements

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/stretchr/testify/assert"
)

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()

		assert.NotNil(t, ProvideService(logger))
	})
}
