package auth_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
)

func TestProvideBcryptHashCost(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		auth.ProvideBcryptHashCost()
	})
}
