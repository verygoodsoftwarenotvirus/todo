package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccountRolePermissionChecker(T *testing.T) {
	T.Parallel()

	T.Run("account user", func(t *testing.T) {
		t.Parallel()

		r := NewAccountRolePermissionChecker(AccountMemberRole.String())

		assert.False(t, r.CanUpdateAccounts())
		assert.False(t, r.CanDeleteAccounts())
		assert.False(t, r.CanAddMemberToAccounts())
		assert.False(t, r.CanRemoveMemberFromAccounts())
		assert.False(t, r.CanTransferAccountToNewOwner())
		assert.False(t, r.CanCreateWebhooks())
		assert.False(t, r.CanSeeWebhooks())
		assert.False(t, r.CanUpdateWebhooks())
		assert.False(t, r.CanArchiveWebhooks())
		assert.False(t, r.CanCreateAPIClients())
		assert.False(t, r.CanSeeAPIClients())
		assert.False(t, r.CanDeleteAPIClients())
		assert.False(t, r.CanSeeAuditLogEntriesForItems())
		assert.False(t, r.CanSeeAuditLogEntriesForWebhooks())
	})

	T.Run("account admin", func(t *testing.T) {
		t.Parallel()

		r := NewAccountRolePermissionChecker(AccountAdminRole.String())

		assert.True(t, r.CanUpdateAccounts())
		assert.True(t, r.CanDeleteAccounts())
		assert.True(t, r.CanAddMemberToAccounts())
		assert.True(t, r.CanRemoveMemberFromAccounts())
		assert.True(t, r.CanTransferAccountToNewOwner())
		assert.True(t, r.CanCreateWebhooks())
		assert.True(t, r.CanSeeWebhooks())
		assert.True(t, r.CanUpdateWebhooks())
		assert.True(t, r.CanArchiveWebhooks())
		assert.True(t, r.CanCreateAPIClients())
		assert.True(t, r.CanSeeAPIClients())
		assert.True(t, r.CanDeleteAPIClients())
		assert.True(t, r.CanSeeAuditLogEntriesForItems())
		assert.True(t, r.CanSeeAuditLogEntriesForWebhooks())
	})
}
