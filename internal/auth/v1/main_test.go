package auth_test

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"testing"
)

func TestProvideBcryptHashCost(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		auth.ProvideBcryptHashCost()
	})

}
