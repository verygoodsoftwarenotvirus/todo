package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceRoles(T *testing.T) {
	T.Parallel()

	T.Run("service user", func(t *testing.T) {
		t.Parallel()

		r := NewServiceRolePermissionChecker(ServiceUserRole.String())

		assert.False(t, r.IsServiceAdmin())
		assert.False(t, r.CanCycleCookieSecrets())
		assert.False(t, r.CanUpdateUserReputations())
		assert.False(t, r.CanSeeUserData())
		assert.False(t, r.CanSearchUsers())
	})

	T.Run("service admin", func(t *testing.T) {
		t.Parallel()

		r := NewServiceRolePermissionChecker(ServiceAdminRole.String())

		assert.True(t, r.IsServiceAdmin())
		assert.True(t, r.CanCycleCookieSecrets())
		assert.True(t, r.CanUpdateUserReputations())
		assert.True(t, r.CanSeeUserData())
		assert.True(t, r.CanSearchUsers())
	})
}
