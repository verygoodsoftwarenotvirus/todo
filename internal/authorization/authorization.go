package authorization

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func hasPermission(p permission, roles ...string) bool {
	for _, r := range roles {
		if !globalAuthorizer.IsGranted(r, p, nil) {
			return false
		}
	}

	return true
}

// CanCycleCookieSecrets returns whether a user can cycle cookie secrets or not.
func CanCycleCookieSecrets(roles ...string) bool {
	return hasPermission(CycleCookieSecretPermission, roles...)
}

// CanSeeAccountAuditLogEntries returns whether a user can view account audit log entries or not.
func CanSeeAccountAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadAccountAuditLogEntriesPermission, roles...)
}

// CanSeeAPIClientAuditLogEntries returns whether a user can view API client audit log entries or not.
func CanSeeAPIClientAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadAPIClientAuditLogEntriesPermission, roles...)
}

// CanSeeUserAuditLogEntries returns whether a user can view user audit log entries or not.
func CanSeeUserAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadUserAuditLogEntriesPermission, roles...)
}

// CanSeeWebhookAuditLogEntries returns whether a user can view webhook audit log entries or not.
func CanSeeWebhookAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadWebhookAuditLogEntriesPermission, roles...)
}

// CanUpdateUserReputations returns whether a user can update user reputations or not.
func CanUpdateUserReputations(roles ...string) bool {
	return hasPermission(UpdateUserReputationPermission, roles...)
}

// CanSeeUserData returns whether a user can view users or not.
func CanSeeUserData(roles ...string) bool {
	return hasPermission(ReadUserPermission, roles...)
}

// CanSearchUsers returns whether a user can search for users or not.
func CanSearchUsers(roles ...string) bool {
	return hasPermission(SearchUserPermission, roles...)
}

// CanUpdateAccounts returns whether a user can update accounts or not.
func CanUpdateAccounts(roles ...string) bool {
	return hasPermission(UpdateAccountPermission, roles...)
}

// CanDeleteAccounts returns whether a user can delete accounts or not.
func CanDeleteAccounts(roles ...string) bool {
	return hasPermission(DeleteAccountPermission, roles...)
}

// CanAddMemberToAccounts returns whether a user can add members to accounts or not.
func CanAddMemberToAccounts(roles ...string) bool {
	return hasPermission(AddMemberAccountPermission, roles...)
}

// CanRemoveMemberFromAccounts returns whether a user can remove members from accounts or not.
func CanRemoveMemberFromAccounts(roles ...string) bool {
	return hasPermission(RemoveMemberAccountPermission, roles...)
}

// CanTransferAccountToNewOwner returns whether a user can transfer an account to a new owner or not.
func CanTransferAccountToNewOwner(roles ...string) bool {
	return hasPermission(TransferAccountPermission, roles...)
}

// CanCreateWebhooks returns whether a user can create webhooks or not.
func CanCreateWebhooks(roles ...string) bool {
	return hasPermission(CreateWebhooksPermission, roles...)
}

// CanSeeWebhooks returns whether a user can view webhooks or not.
func CanSeeWebhooks(roles ...string) bool {
	return hasPermission(ReadWebhooksPermission, roles...)
}

// CanUpdateWebhooks returns whether a user can update webhooks or not.
func CanUpdateWebhooks(roles ...string) bool {
	return hasPermission(UpdateWebhooksPermission, roles...)
}

// CanDeleteWebhooks returns whether a user can delete webhooks or not.
func CanDeleteWebhooks(roles ...string) bool {
	return hasPermission(DeleteWebhooksPermission, roles...)
}

// CanCreateAPIClients returns whether a user can create API clients or not.
func CanCreateAPIClients(roles ...string) bool {
	return hasPermission(CreateAPIClientsPermission, roles...)
}

// CanSeeAPIClients returns whether a user can view API clients or not.
func CanSeeAPIClients(roles ...string) bool {
	return hasPermission(ReadAPIClientsPermission, roles...)
}

// CanDeleteAPIClients returns whether a user can delete API clients or not.
func CanDeleteAPIClients(roles ...string) bool {
	return hasPermission(DeleteAPIClientsPermission, roles...)
}

// CanSeeItemsAuditLogEntries returns whether a user can view item audit log entries or not.
func CanSeeItemsAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadItemsAuditLogEntriesPermission, roles...)
}

// CanCreateItems returns whether a user can create items or not.
func CanCreateItems(roles ...string) bool {
	return hasPermission(CreateItemsPermission, roles...)
}

// CanSeeItems returns whether a user can view items or not.
func CanSeeItems(roles ...string) bool {
	return hasPermission(ReadItemsPermission, roles...)
}

// CanSearchItems returns whether a user can search items or not.
func CanSearchItems(roles ...string) bool {
	return hasPermission(SearchItemsPermission, roles...)
}

// CanUpdateItems returns whether a user can update items or not.
func CanUpdateItems(roles ...string) bool {
	return hasPermission(UpdateItemsPermission, roles...)
}

// CanDeleteItems returns whether a user can delete items or not.
func CanDeleteItems(roles ...string) bool {
	return hasPermission(DeleteItemsPermission, roles...)
}
