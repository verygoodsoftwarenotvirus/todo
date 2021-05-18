package authorization

import (
	"encoding/gob"

	"gopkg.in/mikespook/gorbac.v2"
)

const (
	serviceAdminRoleName = "service_admin"
	serviceUserRoleName  = "service_user"
)

type (
	// ServiceRole describes a role a user has for the Service context.
	ServiceRole role

	// ServiceRolePermissionChecker checks permissions for one or more service Roles.
	ServiceRolePermissionChecker interface {
		IsServiceAdmin() bool
		CanCycleCookieSecrets() bool
		CanSeeAccountAuditLogEntries() bool
		CanSeeAPIClientAuditLogEntries() bool
		CanSeeUserAuditLogEntries() bool
		CanSeeWebhookAuditLogEntries() bool
		CanUpdateUserReputations() bool
		CanSeeUserData() bool
		CanSearchUsers() bool
	}
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

// ServiceRoleFromString returns a service role from a string, or possibly an error for an invalid string.
func ServiceRoleFromString(s string) ServiceRole {
	switch s {
	case serviceAdminRoleName:
		return ServiceAdminRole
	case serviceUserRoleName:
		return ServiceUserRole
	default:
		return invalidServiceRole
	}
}

type serviceRoleCollection struct {
	Roles []string
}

func init() {
	gob.Register(serviceRoleCollection{})
}

// NewServiceRolePermissionChecker returns a new checker for a set of Roles.
func NewServiceRolePermissionChecker(roles ...string) ServiceRolePermissionChecker {
	return &serviceRoleCollection{
		Roles: roles,
	}
}

// IsServiceAdmin returns if a role is an admin. Should probably be DELTEME'd.
func (r serviceRoleCollection) IsServiceAdmin() bool {
	for _, x := range r.Roles {
		if x == ServiceAdminRole.String() {
			return true
		}
	}

	return false
}

// CanCycleCookieSecrets returns whether a user can cycle cookie secrets or not.
func (r serviceRoleCollection) CanCycleCookieSecrets() bool {
	return hasPermission(CycleCookieSecretPermission, r.Roles...)
}

// CanSeeAccountAuditLogEntries returns whether a user can view account audit log entries or not.
func (r serviceRoleCollection) CanSeeAccountAuditLogEntries() bool {
	return hasPermission(ReadAccountAuditLogEntriesPermission, r.Roles...)
}

// CanSeeAPIClientAuditLogEntries returns whether a user can view API client audit log entries or not.
func (r serviceRoleCollection) CanSeeAPIClientAuditLogEntries() bool {
	return hasPermission(ReadAPIClientAuditLogEntriesPermission, r.Roles...)
}

// CanSeeUserAuditLogEntries returns whether a user can view user audit log entries or not.
func (r serviceRoleCollection) CanSeeUserAuditLogEntries() bool {
	return hasPermission(ReadUserAuditLogEntriesPermission, r.Roles...)
}

// CanSeeWebhookAuditLogEntries returns whether a user can view webhook audit log entries or not.
func (r serviceRoleCollection) CanSeeWebhookAuditLogEntries() bool {
	return hasPermission(ReadWebhookAuditLogEntriesPermission, r.Roles...)
}

// CanUpdateUserReputations returns whether a user can update user reputations or not.
func (r serviceRoleCollection) CanUpdateUserReputations() bool {
	return hasPermission(UpdateUserReputationPermission, r.Roles...)
}

// CanSeeUserData returns whether a user can view users or not.
func (r serviceRoleCollection) CanSeeUserData() bool {
	return hasPermission(ReadUserPermission, r.Roles...)
}

// CanSearchUsers returns whether a user can search for users or not.
func (r serviceRoleCollection) CanSearchUsers() bool {
	return hasPermission(SearchUserPermission, r.Roles...)
}
