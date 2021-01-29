package postgres

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildGetPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.id = $1"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllPlansCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL"
		actualQuery := q.BuildGetAllAccountSubscriptionPlansCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_BuildGetAccountSubscriptionPlansQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL) as total_count, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.created_on > $1 AND account_subscription_plans.created_on < $2 AND account_subscription_plans.last_updated_on > $3 AND account_subscription_plans.last_updated_on < $4) as filtered_count FROM account_subscription_plans WHERE account_subscription_plans.created_on > $5 AND account_subscription_plans.created_on < $6 AND account_subscription_plans.last_updated_on > $7 AND account_subscription_plans.last_updated_on < $8 GROUP BY account_subscription_plans.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildCreateAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		expectedQuery := "INSERT INTO account_subscription_plans (name,description,price,period) VALUES ($1,$2,$3,$4) RETURNING id"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
		}
		actualQuery, actualArgs := q.BuildCreateAccountSubscriptionPlanQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildUpdateAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET name = $1, description = $2, price = $3, period = $4, last_updated_on = extract(epoch FROM NOW()) WHERE id = $5"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateAccountSubscriptionPlanQuery(examplePlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildArchiveAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $1"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.context->'plan_id' = $1 ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
