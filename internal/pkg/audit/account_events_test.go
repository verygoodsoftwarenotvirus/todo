package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleAccountID uint64 = 123
)

func TestAccountEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildAccountCreationEventEntry": {
			expectedEventType: audit.AccountCreationEvent,
			expectedContextKeys: []string{
				audit.UserAssignmentKey,
				audit.AccountAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildAccountCreationEventEntry(&types.Account{}),
		},
		"BuildAccountUpdateEventEntry": {
			expectedEventType: audit.AccountUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.ChangesAssignmentKey,
			},
			actual: audit.BuildAccountUpdateEventEntry(exampleUserID, exampleAccountID, nil),
		},
		"BuildAccountArchiveEventEntry": {
			expectedEventType: audit.AccountArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
			},
			actual: audit.BuildAccountArchiveEventEntry(exampleUserID, exampleAccountID),
		},
	}

	runEventBuilderTests(T, tests)
}
