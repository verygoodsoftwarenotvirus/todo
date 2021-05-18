package authorization

import (
	"gopkg.in/mikespook/gorbac.v2"
)

type (
	role int
)

var (
	globalAuthorizer *gorbac.RBAC
)

func init() {
	globalAuthorizer = initializeRBAC()
}

func initializeRBAC() *gorbac.RBAC {
	rbac := gorbac.New()

	must(rbac.Add(serviceUser))
	must(rbac.Add(serviceAdmin))
	must(rbac.Add(accountAdmin))
	must(rbac.Add(accountMember))

	must(rbac.SetParent(accountAdminRoleName, accountMemberRoleName))
	must(rbac.SetParent(serviceAdminRoleName, accountAdminRoleName))

	return rbac
}
