package items

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideItemDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ProvideItemDataServer(buildTestService())
	})
}

func TestProvideItemsServiceSessionInfoFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideItemsServiceSessionInfoFetcher()
	})
}

func TestProvideItemsServiceItemIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideItemsServiceItemIDFetcher(noop.NewLogger())
	})
}
