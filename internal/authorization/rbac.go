package authorization

import (
	"gopkg.in/mikespook/gorbac.v2"
)

var (
	globalAuthorizer *gorbac.RBAC
)

func init() {
	globalAuthorizer = initializeRBAC()
}

func initializeRBAC() *gorbac.RBAC {
	rbac := gorbac.New()

	must(rbac.Add(serviceAdmin))
	must(rbac.Add(accountAdmin))
	must(rbac.Add(accountMember))

	must(rbac.SetParent(AccountAdminRole.Name(), AccountMemberRole.Name()))
	must(rbac.SetParent(ServiceAdminRole.Name(), AccountAdminRole.Name()))

	return rbac
}
