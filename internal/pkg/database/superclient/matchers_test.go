package superclient

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	These are functions that would be greatly compressed by generic support in the runtime. :)
*/

func matchAuditLogEntry(t *testing.T, expected *types.AuditLogEntry) func(*types.AuditLogEntry) bool {
	t.Helper()

	require.NotNil(t, expected, "Webhook matcher invoked with nil webhook")

	x := &types.AuditLogEntry{}
	*x = *expected

	return func(actual *types.AuditLogEntry) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual webhooks do not match")
	}
}

func matchItem(t *testing.T, expected *types.Item) func(*types.Item) bool {
	t.Helper()

	require.NotNil(t, expected, "Item matcher invoked with nil webhook")

	x := &types.Item{}
	*x = *expected

	return func(actual *types.Item) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual webhooks do not match")
	}
}

func matchWebhook(t *testing.T, expected *types.Webhook) func(*types.Webhook) bool {
	t.Helper()

	require.NotNil(t, expected, "Webhook matcher invoked with nil webhook")

	x := &types.Webhook{}
	*x = *expected

	return func(actual *types.Webhook) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual webhooks do not match")
	}
}
