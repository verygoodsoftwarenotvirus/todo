package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

func TestAccountUserMembershipEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildItemCreationEventEntry": {
			expectedEventType: audit.UserAddedToAccountEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserAddedToAccountEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID, "blah blah"),
		},
		"BuildItemUpdateEventEntry": {
			expectedEventType: audit.AccountUserMembershipArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserRemovedFromAccountEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID, "blah blah"),
		},
		"BuildItemArchiveEventEntry": {
			expectedEventType: audit.AccountMarkedAsDefaultEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserMarkedAccountAsDefaultEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID, "blah blah"),
		},
	}

	runEventBuilderTests(T, tests)
}
