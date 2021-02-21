package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleItemID uint64 = 123
)

func TestItemEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildItemCreationEventEntry": {
			expectedEventType: audit.ItemCreationEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.ItemAssignmentKey,
				audit.CreationAssignmentKey,
				audit.AccountAssignmentKey,
			},
			actual: audit.BuildItemCreationEventEntry(&types.Item{}, exampleAccountID),
		},
		"BuildItemUpdateEventEntry": {
			expectedEventType: audit.ItemUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.ItemAssignmentKey,
				audit.AccountAssignmentKey,
				audit.ChangesAssignmentKey,
			},
			actual: audit.BuildItemUpdateEventEntry(exampleUserID, exampleItemID, exampleAccountID, nil),
		},
		"BuildItemArchiveEventEntry": {
			expectedEventType: audit.ItemArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.ItemAssignmentKey,
			},
			actual: audit.BuildItemArchiveEventEntry(exampleUserID, exampleItemID, exampleAccountID),
		},
	}

	runEventBuilderTests(T, tests)
}
