package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

func TestAdminEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildUserBanEventEntry": {
			expectedEventType: audit.UserBannedEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.UserAssignmentKey,
				audit.ReasonKey,
			},
			actual: audit.BuildUserBanEventEntry(exampleUserID, exampleUserID, "reason"),
		},
	}

	runEventBuilderTests(T, tests)
}
