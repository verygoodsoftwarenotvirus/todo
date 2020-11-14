package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

func TestAuditEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildCycleCookieSecretEvent": {
			expectedEventType: audit.CycleCookieSecretEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildCycleCookieSecretEvent(exampleUserID),
		},
		"BuildSuccessfulLoginEventEntry": {
			expectedEventType: audit.SuccessfulLoginEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildSuccessfulLoginEventEntry(exampleUserID),
		},
		"BuildUnsuccessfulLoginBadPasswordEventEntry": {
			expectedEventType: audit.UnsuccessfulLoginBadPasswordEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleUserID),
		},
		"BuildUnsuccessfulLoginBad2FATokenEventEntry": {
			expectedEventType: audit.UnsuccessfulLoginBad2FATokenEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleUserID),
		},
		"BuildLogoutEventEntry": {
			expectedEventType: audit.LogoutEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildLogoutEventEntry(exampleUserID),
		},
	}

	testEventBuilders(T, tests)
}
