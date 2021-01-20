package superclient

import (
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCountDBRowResponse(count uint64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(count)
}

func newSuccessfulDatabaseResult(returnID uint64) driver.Result {
	return sqlmock.NewResult(int64(returnID), 1)
}

/*
	These are functions that would be greatly compressed by generic support in the runtime. :)
*/

func matchAccount(t *testing.T, expected *types.Account) func(*types.Account) bool {
	t.Helper()

	require.NotNil(t, expected, "Webhook matcher invoked with nil webhook")

	x := &types.Account{}
	*x = *expected

	return func(actual *types.Account) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual Accounts do not match")
	}
}

func matchAuditLogEntry(t *testing.T, expected *types.AuditLogEntry) func(*types.AuditLogEntry) bool {
	t.Helper()

	require.NotNil(t, expected, "Webhook matcher invoked with nil webhook")

	x := &types.AuditLogEntry{}
	*x = *expected

	return func(actual *types.AuditLogEntry) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual AuditLogEntries do not match")
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

		return assert.Equal(t, x, actual, "expected and actual Items do not match")
	}
}

func matchUser(t *testing.T, expected *types.User) func(*types.User) bool {
	t.Helper()

	require.NotNil(t, expected, "User matcher invoked with nil webhook")

	x := &types.User{}
	*x = *expected

	return func(actual *types.User) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual Users do not match")
	}
}

func matchAccountSubscriptionPlan(t *testing.T, expected *types.AccountSubscriptionPlan) func(*types.AccountSubscriptionPlan) bool {
	t.Helper()

	require.NotNil(t, expected, "AccountSubscriptionPlan matcher invoked with nil webhook")

	x := &types.AccountSubscriptionPlan{}
	*x = *expected

	return func(actual *types.AccountSubscriptionPlan) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual AccountSubscriptionPlans do not match")
	}
}

func matchOAuth2Client(t *testing.T, expected *types.OAuth2Client) func(*types.OAuth2Client) bool {
	t.Helper()

	require.NotNil(t, expected, "OAuth2Client matcher invoked with nil webhook")

	x := &types.OAuth2Client{}
	*x = *expected

	return func(actual *types.OAuth2Client) bool {
		x.ID = actual.ID
		x.CreatedOn = actual.CreatedOn

		return assert.Equal(t, x, actual, "expected and actual OAuth2Clients do not match")
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

		return assert.Equal(t, x, actual, "expected and actual Webhooks do not match")
	}
}
