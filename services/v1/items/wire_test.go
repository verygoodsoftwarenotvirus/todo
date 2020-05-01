package items

import (
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
)

func TestProvideItemDataManager(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideItemDataManager(database.BuildMockDatabase())
	})
}

func TestProvideItemDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideItemDataServer(buildTestService())
	})
}
