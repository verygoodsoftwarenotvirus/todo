package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideUserDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ProvideUserDataServer(buildTestService(t)))
	})
}

func TestProvideUsersServiceUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideUsersServiceUserIDFetcher(noop.NewLogger())
	})
}

func TestProvideUsersServiceSessionInfoFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideUsersServiceSessionInfoFetcher()
	})
}
