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
		assert.Equal(t, r.HasPermission(UpdateAccountPermission), r.CanUpdateAccounts())
		assert.False(t, r.CanDeleteAccounts())
		assert.Equal(t, r.HasPermission(ArchiveAccountPermission), r.CanDeleteAccounts())
		assert.False(t, r.CanAddMemberToAccounts())
		assert.Equal(t, r.HasPermission(AddMemberAccountPermission), r.CanAddMemberToAccounts())
		assert.False(t, r.CanRemoveMemberFromAccounts())
		assert.Equal(t, r.HasPermission(RemoveMemberAccountPermission), r.CanRemoveMemberFromAccounts())
		assert.False(t, r.CanTransferAccountToNewOwner())
		assert.Equal(t, r.HasPermission(TransferAccountPermission), r.CanTransferAccountToNewOwner())
		assert.False(t, r.CanCreateWebhooks())
		assert.Equal(t, r.HasPermission(CreateWebhooksPermission), r.CanCreateWebhooks())
		assert.False(t, r.CanSeeWebhooks())
		assert.Equal(t, r.HasPermission(ReadWebhooksPermission), r.CanSeeWebhooks())
		assert.False(t, r.CanUpdateWebhooks())
		assert.Equal(t, r.HasPermission(UpdateWebhooksPermission), r.CanUpdateWebhooks())
		assert.False(t, r.CanArchiveWebhooks())
		assert.Equal(t, r.HasPermission(ArchiveWebhooksPermission), r.CanArchiveWebhooks())
		assert.False(t, r.CanCreateAPIClients())
		assert.Equal(t, r.HasPermission(CreateAPIClientsPermission), r.CanCreateAPIClients())
		assert.False(t, r.CanSeeAPIClients())
		assert.Equal(t, r.HasPermission(ReadAPIClientsPermission), r.CanSeeAPIClients())
		assert.False(t, r.CanDeleteAPIClients())
		assert.Equal(t, r.HasPermission(ArchiveAPIClientsPermission), r.CanDeleteAPIClients())
	})

	T.Run("account admin", func(t *testing.T) {
		t.Parallel()

		r := NewAccountRolePermissionChecker(AccountAdminRole.String())

		assert.True(t, r.CanUpdateAccounts())
		assert.Equal(t, r.HasPermission(UpdateAccountPermission), r.CanUpdateAccounts())
		assert.True(t, r.CanDeleteAccounts())
		assert.Equal(t, r.HasPermission(ArchiveAccountPermission), r.CanDeleteAccounts())
		assert.True(t, r.CanAddMemberToAccounts())
		assert.Equal(t, r.HasPermission(AddMemberAccountPermission), r.CanAddMemberToAccounts())
		assert.True(t, r.CanRemoveMemberFromAccounts())
		assert.Equal(t, r.HasPermission(RemoveMemberAccountPermission), r.CanRemoveMemberFromAccounts())
		assert.True(t, r.CanTransferAccountToNewOwner())
		assert.Equal(t, r.HasPermission(TransferAccountPermission), r.CanTransferAccountToNewOwner())
		assert.True(t, r.CanCreateWebhooks())
		assert.Equal(t, r.HasPermission(CreateWebhooksPermission), r.CanCreateWebhooks())
		assert.True(t, r.CanSeeWebhooks())
		assert.Equal(t, r.HasPermission(ReadWebhooksPermission), r.CanSeeWebhooks())
		assert.True(t, r.CanUpdateWebhooks())
		assert.Equal(t, r.HasPermission(UpdateWebhooksPermission), r.CanUpdateWebhooks())
		assert.True(t, r.CanArchiveWebhooks())
		assert.Equal(t, r.HasPermission(ArchiveWebhooksPermission), r.CanArchiveWebhooks())
		assert.True(t, r.CanCreateAPIClients())
		assert.Equal(t, r.HasPermission(CreateAPIClientsPermission), r.CanCreateAPIClients())
		assert.True(t, r.CanSeeAPIClients())
		assert.Equal(t, r.HasPermission(ReadAPIClientsPermission), r.CanSeeAPIClients())
		assert.True(t, r.CanDeleteAPIClients())
		assert.Equal(t, r.HasPermission(ArchiveAPIClientsPermission), r.CanDeleteAPIClients())
	})
}
