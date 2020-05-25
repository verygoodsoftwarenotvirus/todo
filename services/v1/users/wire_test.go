package users

import (
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
)

func TestProvideReportDataManager(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideUserDataManager(database.BuildMockDatabase())
	})
}

func TestProvideReportDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideUserDataServer(buildTestService(t))
	})
}
