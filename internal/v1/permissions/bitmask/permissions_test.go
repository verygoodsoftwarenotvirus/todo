package bitmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_permissionMask_CanCycleCookieSecrets(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		assert.True(t, NewPermissionMask(1<<32-1).CanCycleCookieSecrets())
	})
}
