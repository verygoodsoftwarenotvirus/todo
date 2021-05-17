package authorization

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthorizations(T *testing.T) {
	T.Parallel()

	T.Run("service admin", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CanSeeAccountAuditLogEntries(ServiceAdminRoleName))
		assert.True(t, CanSeeAPIClientAuditLogEntries(ServiceAdminRoleName))
		assert.True(t, CanSeeUserAuditLogEntries(ServiceAdminRoleName))
		assert.True(t, CanSeeWebhookAuditLogEntries(ServiceAdminRoleName))
		assert.True(t, CanUpdateUserReputations(ServiceAdminRoleName))
		assert.True(t, CanSeeUserData(ServiceAdminRoleName))
		assert.True(t, CanSearchUsers(ServiceAdminRoleName))
		assert.True(t, CanUpdateAccounts(ServiceAdminRoleName))
		assert.True(t, CanDeleteAccounts(ServiceAdminRoleName))
		assert.True(t, CanAddMemberToAccounts(ServiceAdminRoleName))
		assert.True(t, CanRemoveMemberFromAccounts(ServiceAdminRoleName))
		assert.True(t, CanTransferAccountToNewOwner(ServiceAdminRoleName))
		assert.True(t, CanCreateWebhooks(ServiceAdminRoleName))
		assert.True(t, CanSeeWebhooks(ServiceAdminRoleName))
		assert.True(t, CanUpdateWebhooks(ServiceAdminRoleName))
		assert.True(t, CanDeleteWebhooks(ServiceAdminRoleName))
		assert.True(t, CanCreateAPIClients(ServiceAdminRoleName))
		assert.True(t, CanSeeAPIClients(ServiceAdminRoleName))
		assert.True(t, CanDeleteAPIClients(ServiceAdminRoleName))
		assert.True(t, CanSeeItemsAuditLogEntries(ServiceAdminRoleName))
		assert.True(t, CanCreateItems(ServiceAdminRoleName))
		assert.True(t, CanSeeItems(ServiceAdminRoleName))
		assert.True(t, CanSearchItems(ServiceAdminRoleName))
		assert.True(t, CanUpdateItems(ServiceAdminRoleName))
		assert.True(t, CanDeleteItems(ServiceAdminRoleName))
	})

	T.Run("account admin", func(t *testing.T) {
		t.Parallel()

		assert.False(t, CanSeeAccountAuditLogEntries(AccountAdminRoleName))
		assert.False(t, CanSeeAPIClientAuditLogEntries(AccountAdminRoleName))
		assert.False(t, CanSeeUserAuditLogEntries(AccountAdminRoleName))
		assert.False(t, CanSeeWebhookAuditLogEntries(AccountAdminRoleName))
		assert.False(t, CanUpdateUserReputations(AccountAdminRoleName))
		assert.False(t, CanSeeUserData(AccountAdminRoleName))
		assert.False(t, CanSearchUsers(AccountAdminRoleName))
		assert.True(t, CanUpdateAccounts(AccountAdminRoleName))
		assert.True(t, CanDeleteAccounts(AccountAdminRoleName))
		assert.True(t, CanAddMemberToAccounts(AccountAdminRoleName))
		assert.True(t, CanRemoveMemberFromAccounts(AccountAdminRoleName))
		assert.True(t, CanTransferAccountToNewOwner(AccountAdminRoleName))
		assert.True(t, CanCreateWebhooks(AccountAdminRoleName))
		assert.True(t, CanSeeWebhooks(AccountAdminRoleName))
		assert.True(t, CanUpdateWebhooks(AccountAdminRoleName))
		assert.True(t, CanDeleteWebhooks(AccountAdminRoleName))
		assert.True(t, CanCreateAPIClients(AccountAdminRoleName))
		assert.True(t, CanSeeAPIClients(AccountAdminRoleName))
		assert.True(t, CanDeleteAPIClients(AccountAdminRoleName))
		assert.True(t, CanSeeItemsAuditLogEntries(AccountAdminRoleName))
		assert.True(t, CanCreateItems(AccountAdminRoleName))
		assert.True(t, CanSeeItems(AccountAdminRoleName))
		assert.True(t, CanSearchItems(AccountAdminRoleName))
		assert.True(t, CanUpdateItems(AccountAdminRoleName))
		assert.True(t, CanDeleteItems(AccountAdminRoleName))
	})

	T.Run("account member", func(t *testing.T) {
		t.Parallel()

		assert.False(t, CanSeeAccountAuditLogEntries(AccountMemberRoleName))
		assert.False(t, CanSeeAPIClientAuditLogEntries(AccountMemberRoleName))
		assert.False(t, CanSeeUserAuditLogEntries(AccountMemberRoleName))
		assert.False(t, CanSeeWebhookAuditLogEntries(AccountMemberRoleName))
		assert.False(t, CanUpdateUserReputations(AccountMemberRoleName))
		assert.False(t, CanSeeUserData(AccountMemberRoleName))
		assert.False(t, CanSearchUsers(AccountMemberRoleName))
		assert.False(t, CanUpdateAccounts(AccountMemberRoleName))
		assert.False(t, CanDeleteAccounts(AccountMemberRoleName))
		assert.False(t, CanAddMemberToAccounts(AccountMemberRoleName))
		assert.False(t, CanRemoveMemberFromAccounts(AccountMemberRoleName))
		assert.False(t, CanTransferAccountToNewOwner(AccountMemberRoleName))
		assert.False(t, CanCreateWebhooks(AccountMemberRoleName))
		assert.False(t, CanSeeWebhooks(AccountMemberRoleName))
		assert.False(t, CanUpdateWebhooks(AccountMemberRoleName))
		assert.False(t, CanDeleteWebhooks(AccountMemberRoleName))
		assert.False(t, CanCreateAPIClients(AccountMemberRoleName))
		assert.False(t, CanSeeAPIClients(AccountMemberRoleName))
		assert.False(t, CanDeleteAPIClients(AccountMemberRoleName))
		assert.False(t, CanSeeItemsAuditLogEntries(AccountMemberRoleName))
		assert.True(t, CanCreateItems(AccountMemberRoleName))
		assert.True(t, CanSeeItems(AccountMemberRoleName))
		assert.True(t, CanSearchItems(AccountMemberRoleName))
		assert.True(t, CanUpdateItems(AccountMemberRoleName))
		assert.True(t, CanDeleteItems(AccountMemberRoleName))
	})
}
