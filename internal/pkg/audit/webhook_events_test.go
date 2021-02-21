package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleWebhookID uint64 = 123
)

func TestWebhookEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildWebhookCreationEventEntry": {
			expectedEventType: audit.WebhookCreationEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.CreationAssignmentKey,
				audit.WebhookAssignmentKey,
				audit.AccountAssignmentKey,
			},
			actual: audit.BuildWebhookCreationEventEntry(&types.Webhook{}, exampleUserID),
		},
		"BuildWebhookUpdateEventEntry": {
			expectedEventType: audit.WebhookUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.WebhookAssignmentKey,
				audit.ChangesAssignmentKey,
				audit.AccountAssignmentKey,
			},
			actual: audit.BuildWebhookUpdateEventEntry(exampleUserID, exampleAccountID, exampleWebhookID, nil),
		},
		"BuildWebhookArchiveEventEntry": {
			expectedEventType: audit.WebhookArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.WebhookAssignmentKey,
				audit.UserAssignmentKey,
			},
			actual: audit.BuildWebhookArchiveEventEntry(exampleUserID, exampleAccountID, exampleWebhookID),
		},
	}

	runEventBuilderTests(T, tests)
}
