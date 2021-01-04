package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleAccountSubscriptionPlanID uint64 = 123
)

func TestAccountSubscriptionPlanEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildAccountSubscriptionPlanCreationEventEntry": {
			expectedEventType: audit.AccountSubscriptionPlanCreationEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountSubscriptionPlanAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildAccountSubscriptionPlanCreationEventEntry(&types.AccountSubscriptionPlan{}),
		},
		"BuildAccountSubscriptionPlanUpdateEventEntry": {
			expectedEventType: audit.AccountSubscriptionPlanUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountSubscriptionPlanAssignmentKey,
				audit.ChangesAssignmentKey,
			},
			actual: audit.BuildAccountSubscriptionPlanUpdateEventEntry(exampleUserID, exampleAccountSubscriptionPlanID, nil),
		},
		"BuildAccountSubscriptionPlanArchiveEventEntry": {
			expectedEventType: audit.AccountSubscriptionPlanArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountSubscriptionPlanAssignmentKey,
			},
			actual: audit.BuildAccountSubscriptionPlanArchiveEventEntry(exampleUserID, exampleAccountSubscriptionPlanID),
		},
	}

	runEventBuilderTests(T, tests)
}
