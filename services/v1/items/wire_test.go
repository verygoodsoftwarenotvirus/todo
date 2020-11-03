package items

import (
	"testing"
)

func TestProvideItemDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ProvideItemDataServer(buildTestService())
	})
}
