package sqlite

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildGetPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.id = ?"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAllPlansCountQuery(T *testing.T) {
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

func TestSqlite_BuildGetPlansQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL) as total_count, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.created_on > ? AND account_subscription_plans.created_on < ? AND account_subscription_plans.last_updated_on > ? AND account_subscription_plans.last_updated_on < ?) as filtered_count FROM account_subscription_plans WHERE account_subscription_plans.created_on > ? AND account_subscription_plans.created_on < ? AND account_subscription_plans.last_updated_on > ? AND account_subscription_plans.last_updated_on < ? GROUP BY account_subscription_plans.id LIMIT 20 OFFSET 180"
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

func TestSqlite_BuildCreatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInput()

		expectedQuery := "INSERT INTO account_subscription_plans (name,description,price,period) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleInput.Name,
			exampleInput.Description,
			exampleInput.Price,
			exampleInput.Period.String(),
		}
		actualQuery, actualArgs := q.BuildCreateAccountSubscriptionPlanQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildUpdatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET name = ?, description = ?, price = ?, period = ?, last_updated_on = (strftime('%s','now')) WHERE id = ?"
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

func TestSqlite_BuildArchivePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		expectedQuery := "UPDATE account_subscription_plans SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
