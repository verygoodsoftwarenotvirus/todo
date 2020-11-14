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
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildItemCreationEventEntry(&types.Item{}),
		},
		"BuildItemUpdateEventEntry": {
			expectedEventType: audit.ItemUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.ItemAssignmentKey,
				audit.ChangesAssignmentKey,
			},
			actual: audit.BuildItemUpdateEventEntry(exampleUserID, exampleItemID, nil),
		},
		"BuildItemArchiveEventEntry": {
			expectedEventType: audit.ItemArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.ItemAssignmentKey,
			},
			actual: audit.BuildItemArchiveEventEntry(exampleUserID, exampleItemID),
		},
	}

	testEventBuilders(T, tests)
}
