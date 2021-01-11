package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkPlanEquality(t *testing.T, expected, actual *types.AccountSubscriptionPlan) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for account subscription plan #%d to be %q, but it was %q ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description, "expected Description for account subscription plan #%d to be %q, but it was %q ", expected.ID, expected.Description, actual.Description)
	assert.Equal(t, expected.Price, actual.Price, "expected Price for account subscription plan #%d to be %v, but it was %v ", expected.ID, expected.Price, actual.Price)
	assert.Equal(t, expected.Period, actual.Period, "expected Period for account subscription plan #%d to be %v, but it was %v ", expected.ID, expected.Period, actual.Period)
	assert.NotZero(t, actual.CreatedOn)
}

func TestAccountSubscriptionPlans(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be createable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Create plan.
			examplePlan := fakes.BuildFakePlan()
			examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdPlan, err := adminClient.CreatePlan(ctx, examplePlanInput)
			checkValueAndError(t, createdPlan, err)

			// Assert plan equality.
			checkPlanEquality(t, examplePlan, createdPlan)

			// Clean up.
			err = adminClient.ArchivePlan(ctx, createdPlan.ID)
			assert.NoError(t, err)

			auditLogEntries, err := adminClient.GetAuditLogForPlan(ctx, createdPlan.ID)

			require.Len(t, auditLogEntries, 2)
			require.NoError(t, err)

			expectedEventTypes := []string{
				audit.AccountSubscriptionPlanCreationEvent,
				audit.AccountSubscriptionPlanArchiveEvent,
			}
			actualEventTypes := []string{}

			for _, e := range auditLogEntries {
				actualEventTypes = append(actualEventTypes, e.EventType)
				require.Contains(t, e.Context, audit.AccountSubscriptionPlanAssignmentKey)
				assert.EqualValues(t, createdPlan.ID, e.Context[audit.AccountSubscriptionPlanAssignmentKey])
			}

			assert.Subset(t, expectedEventTypes, actualEventTypes)
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Create plans.
			var created []*types.AccountSubscriptionPlan
			for i := 0; i < 5; i++ {
				// Create plan.
				examplePlan := fakes.BuildFakePlan()
				examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)
				createdPlan, planCreationErr := adminClient.CreatePlan(ctx, examplePlanInput)
				checkValueAndError(t, createdPlan, planCreationErr)

				created = append(created, createdPlan)
			}

			// Assert plan list equality.
			actual, err := adminClient.GetPlans(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(created) <= len(actual.Plans),
				"created %d to be <= %d",
				len(created),
				len(actual.Plans),
			)

			// Clean up.
			for _, plan := range created {
				assert.NoError(t, adminClient.ArchivePlan(ctx, plan.ID))
			}
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Attempt to fetch nonexistent plan.
			_, err := adminClient.GetPlan(ctx, nonexistentID)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Create plan.
			examplePlan := fakes.BuildFakePlan()
			examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdPlan, err := adminClient.CreatePlan(ctx, examplePlanInput)
			checkValueAndError(t, createdPlan, err)

			// Fetch plan.
			actual, err := adminClient.GetPlan(ctx, createdPlan.ID)
			checkValueAndError(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, examplePlan, actual)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchivePlan(ctx, createdPlan.ID))
		})
	})

	test.Run("Updating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			examplePlan := fakes.BuildFakePlan()
			examplePlan.ID = nonexistentID

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			assert.Error(t, adminClient.UpdatePlan(ctx, examplePlan))
		})

		subtest.Run("it should be updatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Create plan.
			examplePlan := fakes.BuildFakePlan()
			examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdPlan, err := adminClient.CreatePlan(ctx, examplePlanInput)
			checkValueAndError(t, createdPlan, err)

			// Change plan.
			createdPlan.Update(converters.ConvertPlanToPlanUpdateInput(examplePlan))
			assert.NoError(t, adminClient.UpdatePlan(ctx, createdPlan))

			// Fetch plan.
			actual, err := adminClient.GetPlan(ctx, createdPlan.ID)
			checkValueAndError(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, examplePlan, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchivePlan(ctx, createdPlan.ID))
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to delete something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			assert.Error(t, adminClient.ArchivePlan(ctx, nonexistentID))
		})

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Create plan.
			examplePlan := fakes.BuildFakePlan()
			examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdPlan, err := adminClient.CreatePlan(ctx, examplePlanInput)
			checkValueAndError(t, createdPlan, err)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchivePlan(ctx, createdPlan.ID))
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			examplePlan := fakes.BuildFakePlan()
			examplePlan.ID = nonexistentID

			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			x, err := adminClient.GetAuditLogForPlan(ctx, examplePlan.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Create plan.
			examplePlan := fakes.BuildFakePlan()
			examplePlanInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)
			createdPlan, err := adminClient.CreatePlan(ctx, examplePlanInput)
			checkValueAndError(t, createdPlan, err)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForPlan(ctx, createdPlan.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchivePlan(ctx, createdPlan.ID))
		})
	})
}
