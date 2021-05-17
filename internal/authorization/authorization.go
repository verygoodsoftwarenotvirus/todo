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

func CanCycleCookieSecrets(roles ...string) bool {
	return hasPermission(CycleCookieSecretPermission, roles...)
}

func CanSeeAccountAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadAccountAuditLogEntriesPermission, roles...)
}

func CanSeeAPIClientAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadAPIClientAuditLogEntriesPermission, roles...)
}

func CanSeeUserAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadUserAuditLogEntriesPermission, roles...)
}

func CanSeeWebhookAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadWebhookAuditLogEntriesPermission, roles...)
}

func CanUpdateUserReputations(roles ...string) bool {
	return hasPermission(UpdateUserReputationPermission, roles...)
}

func CanSeeUserData(roles ...string) bool {
	return hasPermission(ReadUserPermission, roles...)
}

func CanSearchUsers(roles ...string) bool {
	return hasPermission(SearchUserPermission, roles...)
}

func CanUpdateAccounts(roles ...string) bool {
	return hasPermission(UpdateAccountPermission, roles...)
}

func CanDeleteAccounts(roles ...string) bool {
	return hasPermission(DeleteAccountPermission, roles...)
}

func CanAddMemberToAccounts(roles ...string) bool {
	return hasPermission(AddMemberAccountPermission, roles...)
}

func CanRemoveMemberFromAccounts(roles ...string) bool {
	return hasPermission(RemoveMemberAccountPermission, roles...)
}

func CanTransferAccountToNewOwner(roles ...string) bool {
	return hasPermission(TransferAccountPermission, roles...)
}

func CanCreateWebhooks(roles ...string) bool {
	return hasPermission(CreateWebhooksPermission, roles...)
}

func CanSeeWebhooks(roles ...string) bool {
	return hasPermission(ReadWebhooksPermission, roles...)
}

func CanUpdateWebhooks(roles ...string) bool {
	return hasPermission(UpdateWebhooksPermission, roles...)
}

func CanDeleteWebhooks(roles ...string) bool {
	return hasPermission(DeleteWebhooksPermission, roles...)
}

func CanCreateAPIClients(roles ...string) bool {
	return hasPermission(CreateAPIClientsPermission, roles...)
}

func CanSeeAPIClients(roles ...string) bool {
	return hasPermission(ReadAPIClientsPermission, roles...)
}

func CanDeleteAPIClients(roles ...string) bool {
	return hasPermission(DeleteAPIClientsPermission, roles...)
}

func CanSeeItemsAuditLogEntries(roles ...string) bool {
	return hasPermission(ReadItemsAuditLogEntriesPermission, roles...)
}

func CanCreateItems(roles ...string) bool {
	return hasPermission(CreateItemsPermission, roles...)
}

func CanSeeItems(roles ...string) bool {
	return hasPermission(ReadItemsPermission, roles...)
}

func CanSearchItems(roles ...string) bool {
	return hasPermission(SearchItemsPermission, roles...)
}

func CanUpdateItems(roles ...string) bool {
	return hasPermission(UpdateItemsPermission, roles...)
}

func CanDeleteItems(roles ...string) bool {
	return hasPermission(DeleteItemsPermission, roles...)
}
