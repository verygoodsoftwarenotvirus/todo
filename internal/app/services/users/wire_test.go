package users

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

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
