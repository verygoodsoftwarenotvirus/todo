package authorization

import (
	"errors"
	"gopkg.in/mikespook/gorbac.v2"
)

const (
	serviceAdminRoleName = "service_admin"
	serviceUserRoleName  = "service_user"
)

type (
	// ServiceRole describes a role a user has for the Service context.
	ServiceRole role
)

const (
	// invalidServiceRole is a service role to apply for non-admin users to have one.
	invalidServiceRole ServiceRole = iota
	// ServiceUserRole is a service role to apply for non-admin users to have one.
	ServiceUserRole ServiceRole = iota
	// ServiceAdminRole is a role that allows a user to do basically anything.
	ServiceAdminRole ServiceRole = iota
)

var (
	serviceUser  = gorbac.NewStdRole(serviceUserRoleName)
	serviceAdmin = gorbac.NewStdRole(serviceAdminRoleName)
)

func (r ServiceRole) String() string {
	switch r {
	case invalidServiceRole:
		return "INVALID_SERVICE_ROLE"
	case ServiceUserRole:
		return serviceUserRoleName
	case ServiceAdminRole:
		return serviceAdminRoleName
	default:
		return ""
	}
}

var errInvalidServiceRole = errors.New("invalid service role")

// ServiceRoleFromString returns a service role from a string, or possibly an error for an invalid string.
func ServiceRoleFromString(s string) (ServiceRole, error) {
	switch s {
	case serviceAdminRoleName:
		return ServiceAdminRole, nil
	case serviceUserRoleName:
		return ServiceUserRole, nil
	default:
		return invalidServiceRole, errInvalidServiceRole
	}
}

// IsServiceAdmin should be deleted TODO:
func (r ServiceRole) IsServiceAdmin() bool {
	return r == ServiceAdminRole
}

// CanCycleCookieSecrets returns whether a user can cycle cookie secrets or not.
func (r ServiceRole) CanCycleCookieSecrets() bool {
	return hasPermission(CycleCookieSecretPermission, r.String())
}

// CanSeeAccountAuditLogEntries returns whether a user can view account audit log entries or not.
func (r ServiceRole) CanSeeAccountAuditLogEntries() bool {
	return hasPermission(ReadAccountAuditLogEntriesPermission, r.String())
}

// CanSeeAPIClientAuditLogEntries returns whether a user can view API client audit log entries or not.
func (r ServiceRole) CanSeeAPIClientAuditLogEntries() bool {
	return hasPermission(ReadAPIClientAuditLogEntriesPermission, r.String())
}

// CanSeeUserAuditLogEntries returns whether a user can view user audit log entries or not.
func (r ServiceRole) CanSeeUserAuditLogEntries() bool {
	return hasPermission(ReadUserAuditLogEntriesPermission, r.String())
}

// CanSeeWebhookAuditLogEntries returns whether a user can view webhook audit log entries or not.
func (r ServiceRole) CanSeeWebhookAuditLogEntries() bool {
	return hasPermission(ReadWebhookAuditLogEntriesPermission, r.String())
}

// CanUpdateUserReputations returns whether a user can update user reputations or not.
func (r ServiceRole) CanUpdateUserReputations() bool {
	return hasPermission(UpdateUserReputationPermission, r.String())
}

// CanSeeUserData returns whether a user can view users or not.
func (r ServiceRole) CanSeeUserData() bool {
	return hasPermission(ReadUserPermission, r.String())
}

// CanSearchUsers returns whether a user can search for users or not.
func (r ServiceRole) CanSearchUsers() bool {
	return hasPermission(SearchUserPermission, r.String())
}
