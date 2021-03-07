package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

func TestAccountUserMembershipEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildUserAddedToAccountEventEntry": {
			expectedEventType: audit.UserAddedToAccountEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.PermissionsKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserAddedToAccountEventEntry(exampleAdminUserID, exampleAccountID, &types.AddUserToAccountInput{}),
		},
		"BuildUserRemovedFromAccountEventEntry": {
			expectedEventType: audit.UserRemovedFromAccountEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserRemovedFromAccountEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID, "blah blah"),
		},
		"BuildUserMarkedAccountAsDefaultEventEntry": {
			expectedEventType: audit.AccountMarkedAsDefaultEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserMarkedAccountAsDefaultEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID),
		},
	}

	runEventBuilderTests(T, tests)
}
