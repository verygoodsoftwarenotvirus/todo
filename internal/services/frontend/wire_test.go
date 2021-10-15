package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
)

func TestProvideAuthService(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ProvideAuthService(&mocktypes.AuthService{}))
	})
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ProvideUsersService(&mocktypes.UsersService{}))
	})
}
