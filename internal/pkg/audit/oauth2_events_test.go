package audit_test

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"testing"
)

const (
	exampleOAuth2ClientID uint64 = 123
)

func TestOAuth2EventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]eventBuilderTest{
		"BuildOAuth2ClientCreationEventEntry": {
			expectedEventType: audit.OAuth2ClientCreationEvent,
			expectedContextKeys: []string{
				audit.OAuth2ClientAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildOAuth2ClientCreationEventEntry(&types.OAuth2Client{}),
		},
		"BuildOAuth2ClientArchiveEventEntry": {
			expectedEventType: audit.OAuth2ClientArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.OAuth2ClientAssignmentKey,
			},
			actual: audit.BuildOAuth2ClientArchiveEventEntry(exampleUserID, exampleOAuth2ClientID),
		},
	}

	testEventBuilders(T, tests)
}
