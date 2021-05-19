package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizations(T *testing.T) {
	T.Parallel()

	T.Run("service user", func(t *testing.T) {
		t.Parallel()

		assert.False(t, CanCreateItems(serviceUserRoleName))
		assert.False(t, CanSeeItems(serviceUserRoleName))
		assert.False(t, CanSearchItems(serviceUserRoleName))
		assert.False(t, CanUpdateItems(serviceUserRoleName))
		assert.False(t, CanDeleteItems(serviceUserRoleName))
	})

	T.Run("service admin", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CanCreateItems(serviceAdminRoleName))
		assert.True(t, CanSeeItems(serviceAdminRoleName))
		assert.True(t, CanSearchItems(serviceAdminRoleName))
		assert.True(t, CanUpdateItems(serviceAdminRoleName))
		assert.True(t, CanDeleteItems(serviceAdminRoleName))
	})

	T.Run("account admin", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CanCreateItems(accountAdminRoleName))
		assert.True(t, CanSeeItems(accountAdminRoleName))
		assert.True(t, CanSearchItems(accountAdminRoleName))
		assert.True(t, CanUpdateItems(accountAdminRoleName))
		assert.True(t, CanDeleteItems(accountAdminRoleName))
	})

	T.Run("account member", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CanCreateItems(accountMemberRoleName))
		assert.True(t, CanSeeItems(accountMemberRoleName))
		assert.True(t, CanSearchItems(accountMemberRoleName))
		assert.True(t, CanUpdateItems(accountMemberRoleName))
		assert.True(t, CanDeleteItems(accountMemberRoleName))
	})
}
