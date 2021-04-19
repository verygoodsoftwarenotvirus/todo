package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleWebhookID uint64 = 123
)

func TestBuildWebhookCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildWebhookCreationEventEntry(&types.Webhook{}, exampleUserID))
}
func TestBuildWebhookUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildWebhookUpdateEventEntry(exampleUserID, exampleAccountID, exampleWebhookID, nil))
}
func TestBuildWebhookArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildWebhookArchiveEventEntry(exampleUserID, exampleAccountID, exampleWebhookID))
}
