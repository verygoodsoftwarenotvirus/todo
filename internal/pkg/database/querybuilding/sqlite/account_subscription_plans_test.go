package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildGetAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.external_id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.id = ?"
		expectedArgs := []interface{}{
			exampleAccountSubscriptionPlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlanQuery(ctx, exampleAccountSubscriptionPlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAllAccountSubscriptionPlansCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL"
		actualQuery := q.BuildGetAllAccountSubscriptionPlansCountQuery(ctx)

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_BuildGetAccountSubscriptionPlansQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.external_id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL) as total_count, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.created_on > ? AND account_subscription_plans.created_on < ? AND account_subscription_plans.last_updated_on > ? AND account_subscription_plans.last_updated_on < ?) as filtered_count FROM account_subscription_plans WHERE account_subscription_plans.created_on > ? AND account_subscription_plans.created_on < ? AND account_subscription_plans.last_updated_on > ? AND account_subscription_plans.last_updated_on < ? GROUP BY account_subscription_plans.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlansQuery(ctx, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildCreateAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleAccountSubscriptionPlan.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO account_subscription_plans (external_id,name,description,price,period) VALUES (?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleAccountSubscriptionPlan.ExternalID,
			exampleAccountSubscriptionPlan.Name,
			exampleAccountSubscriptionPlan.Description,
			exampleAccountSubscriptionPlan.Price,
			exampleAccountSubscriptionPlan.Period.String(),
		}
		actualQuery, actualArgs := q.BuildCreateAccountSubscriptionPlanQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestSqlite_BuildUpdateAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET name = ?, description = ?, price = ?, period = ?, last_updated_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			exampleAccountSubscriptionPlan.Name,
			exampleAccountSubscriptionPlan.Description,
			exampleAccountSubscriptionPlan.Price,
			exampleAccountSubscriptionPlan.Period.String(),
			exampleAccountSubscriptionPlan.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateAccountSubscriptionPlanQuery(ctx, exampleAccountSubscriptionPlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildArchiveAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			exampleAccountSubscriptionPlan.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountSubscriptionPlanQuery(ctx, exampleAccountSubscriptionPlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE json_extract(audit_log.context, '$.plan_id') = ? ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleAccountSubscriptionPlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(ctx, exampleAccountSubscriptionPlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
