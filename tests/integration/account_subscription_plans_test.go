package integration

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
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

func (s *TestSuite) TestAccountSubscriptionPlans_Creating() {
	s.runForEachClientExcept("should be creatable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			createdAccountSubscriptionPlan, err := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			requireNotNilAndNoProblems(t, createdAccountSubscriptionPlan, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, createdAccountSubscriptionPlan)

			auditLogEntries, err := testClients.admin.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)

			// Clean up.
			assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Listing() {
	s.runForEachClientExcept("should be possible to be read in a list", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plans.
			var created []*types.AccountSubscriptionPlan
			for i := 0; i < 5; i++ {
				// Create plan.
				exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
				exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)
				createdPlan, planCreationErr := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
				requireNotNilAndNoProblems(t, createdPlan, planCreationErr)

				created = append(created, createdPlan)
			}

			// Assert plan list equality.
			actual, err := testClients.admin.GetAccountSubscriptionPlans(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(created) <= len(actual.AccountSubscriptionPlans),
				"created %d to be <= %d",
				len(created),
				len(actual.AccountSubscriptionPlans),
			)

			// Clean up.
			for _, plan := range created {
				assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, plan.ID))
			}
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Reading_Returns404ForNonexistentAccountSubscriptionPlan() {
	s.runForEachClientExcept("should fail to read nonexistent plan", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Attempt to fetch nonexistent plan.
			_, err := testClients.admin.GetAccountSubscriptionPlan(ctx, nonexistentID)
			assert.Error(t, err)
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Reading() {
	s.runForEachClientExcept("should be able to be read", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			createdPlan, err := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			requireNotNilAndNoProblems(t, createdPlan, err)

			// Fetch plan.
			actual, err := testClients.admin.GetAccountSubscriptionPlan(ctx, createdPlan.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, actual)

			// Clean up plan.
			assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, createdPlan.ID))
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Updating_Returns404ForNonexistentAccountSubscriptionPlan() {
	s.runForEachClientExcept("should fail to update a non-existent plan", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlan.ID = nonexistentID

			assert.Error(t, testClients.admin.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan))
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Updating() {
	s.runForEachClientExcept("should be able to be updated", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			createdAccountSubscriptionPlan, err := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			requireNotNilAndNoProblems(t, createdAccountSubscriptionPlan, err)

			// Change plan.
			createdAccountSubscriptionPlan.Update(converters.ConvertAccountSubscriptionPlanToPlanUpdateInput(exampleAccountSubscriptionPlan))
			assert.NoError(t, testClients.admin.UpdateAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan))

			// Fetch plan.
			actual, err := testClients.admin.GetAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert plan equality.
			checkPlanEquality(t, exampleAccountSubscriptionPlan, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := testClients.admin.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
				{EventType: audit.AccountSubscriptionPlanUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)

			// Clean up plan.
			assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Archiving_Returns404ForNonexistentAccountSubscriptionPlan() {
	s.runForEachClientExcept("should fail to archive nonexistent plan", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, nonexistentID))
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Archiving() {
	s.runForEachClientExcept("should be possible to archive plan", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

			createdAccountSubscriptionPlan, err := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			requireNotNilAndNoProblems(t, createdAccountSubscriptionPlan, err)

			// Clean up plan.
			assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForAccountSubscriptionPlan(ctx, createdAccountSubscriptionPlan.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountSubscriptionPlanCreationEvent},
				{EventType: audit.AccountSubscriptionPlanArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccountSubscriptionPlan.ID, audit.AccountSubscriptionPlanAssignmentKey)
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Auditing_Returns404ForNonexistentAccountSubscriptionPlan() {
	s.runForEachClientExcept("should fail to audit plan that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			x, err := testClients.admin.GetAuditLogForAccountSubscriptionPlan(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		}
	})
}

func (s *TestSuite) TestAccountSubscriptionPlans_Auditing() {
	s.runForEachClientExcept("should not be auditable by non-admin", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create plan.
			exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
			exampleAccountSubscriptionPlanInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)
			createdPlan, err := testClients.admin.CreateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlanInput)
			requireNotNilAndNoProblems(t, createdPlan, err)

			// attempt to fetch audit log entries
			actual, err := testClients.main.GetAuditLogForAccountSubscriptionPlan(ctx, createdPlan.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up plan.
			assert.NoError(t, testClients.admin.ArchiveAccountSubscriptionPlan(ctx, createdPlan.ID))
		}
	})
}
