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

		subtest.Run("should be creatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdAccountSubscriptionPlan, err := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			checkValueAndError(t, createdAccountSubscriptionPlan, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, createdAccountSubscriptionPlan)

			auditLogEntries, err := adminClient.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)

			// Clean up.
			assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Create plans.
			var created []*types.AccountSubscriptionPlan
			for i := 0; i < 5; i++ {
				// Create plan.
				exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
				exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)
				createdPlan, planCreationErr := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
				checkValueAndError(t, createdPlan, planCreationErr)

				created = append(created, createdPlan)
			}

			// Assert plan list equality.
			actual, err := adminClient.GetAccountSubscriptionPlans(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(created) <= len(actual.AccountSubscriptionPlans),
				"created %d to be <= %d",
				len(created),
				len(actual.AccountSubscriptionPlans),
			)

			// Clean up.
			for _, plan := range created {
				assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, plan.ID))
			}
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Attempt to fetch nonexistent plan.
			_, err := adminClient.GetAccountSubscriptionPlan(ctx, nonexistentID)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdPlan, err := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			checkValueAndError(t, createdPlan, err)

			// Fetch plan.
			actual, err := adminClient.GetAccountSubscriptionPlan(ctx, createdPlan.ID)
			checkValueAndError(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, actual)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, createdPlan.ID))
		})
	})

	test.Run("Updating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlan.ID = nonexistentID

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			assert.Error(t, adminClient.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan))
		})

		subtest.Run("it should be updatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdAccountSubscriptionPlan, err := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			checkValueAndError(t, createdAccountSubscriptionPlan, err)

			// Change plan.
			createdAccountSubscriptionPlan.Update(converters.ConvertAccountSubscriptionPlanToPlanUpdateInput(exampleAccountSubscriptionPlan))
			assert.NoError(t, adminClient.UpdateAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan))

			// Fetch plan.
			actual, err := adminClient.GetAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			checkValueAndError(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := adminClient.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
				{EventType: audit.AccountSubscriptionPlanUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to delete something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			assert.Error(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, nonexistentID))
		})

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			createdAccountSubscriptionPlan, err := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			checkValueAndError(t, createdAccountSubscriptionPlan, err)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))

			auditLogEntries, err := adminClient.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
				{EventType: audit.AccountSubscriptionPlanArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			x, err := adminClient.GetAuditLogForAccountSubscriptionPlan(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			_, _, testClient := createUserAndClientForTest(ctx, t)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)
			createdPlan, err := adminClient.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			checkValueAndError(t, createdPlan, err)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForAccountSubscriptionPlan(ctx, createdPlan.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up plan.
			assert.NoError(t, adminClient.ArchiveAccountSubscriptionPlan(ctx, createdPlan.ID))
		})
	})
}
