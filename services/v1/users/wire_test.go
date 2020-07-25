package users

import (
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"github.com/stretchr/testify/assert"
)

func TestProvideUserDataManager(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		assert.NotNil(t, ProvideUserDataManager(database.BuildMockDatabase()))
	})
}

func TestProvideUserDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		assert.NotNil(t, ProvideUserDataServer(buildTestService(t)))
	})
}
