package audit_test

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"testing"
)

const (
	exampleWebhookID uint64 = 123
)

func TestWebhookEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]eventBuilderTest{
		"BuildWebhookCreationEventEntry": {
			expectedEventType: audit.WebhookCreationEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildWebhookCreationEventEntry(&types.Webhook{}),
		},
		"BuildWebhookUpdateEventEntry": {
			expectedEventType: audit.WebhookUpdateEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.WebhookAssignmentKey,
				audit.ChangesAssignmentKey,
			},
			actual: audit.BuildWebhookUpdateEventEntry(exampleUserID, exampleWebhookID, nil),
		},
		"BuildWebhookArchiveEventEntry": {
			expectedEventType: audit.WebhookArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.WebhookAssignmentKey,
			},
			actual: audit.BuildWebhookArchiveEventEntry(exampleUserID, exampleWebhookID),
		},
	}

	testEventBuilders(T, tests)
}
