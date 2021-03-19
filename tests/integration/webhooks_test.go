package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func checkWebhookEquality(t *testing.T, expected, actual *types.Webhook) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ContentType, actual.ContentType)
	assert.Equal(t, expected.URL, actual.URL)
	assert.Equal(t, expected.Method, actual.Method)
	assert.NotZero(t, actual.CreatedOn)
}

func (s *TestSuite) TestWebhooksCreating() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be createable via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhook, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, createdWebhook, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, createdWebhook)

			actual, err := testClients.main.GetWebhook(ctx, createdWebhook.ID)
			requireNotNilAndNoProblems(t, actual, err)
			checkWebhookEquality(t, exampleWebhook, actual)

			auditLogEntries, err := testClients.admin.GetAuditLogForWebhook(ctx, createdWebhook.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.WebhookCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdWebhook.ID, audit.WebhookAssignmentKey)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))
		})
	}
}

func (s *TestSuite) TestWebhooksReading() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should fail to read non-existent webhook via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Fetch webhook.
			_, err := testClients.main.GetWebhook(ctx, nonexistentID)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be read via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, premade, err)

			// Fetch webhook.
			actual, err := testClients.main.GetWebhook(ctx, premade.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, actual)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, actual.ID))
		})
	}
}

func (s *TestSuite) TestWebhooksListing() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be read in a list via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhooks.
			var expected []*types.Webhook
			for i := 0; i < 5; i++ {
				exampleWebhook := fakes.BuildFakeWebhook()
				exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
				createdWebhook, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
				requireNotNilAndNoProblems(t, createdWebhook, err)

				expected = append(expected, createdWebhook)
			}

			// Assert webhook list equality.
			actual, err := testClients.main.GetWebhooks(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Webhooks))

			// Clean up.
			for _, webhook := range actual.Webhooks {
				assert.NoError(t, testClients.main.ArchiveWebhook(ctx, webhook.ID))
			}
		})
	}
}

func (s *TestSuite) TestWebhooksUpdating() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should fail to update a non-existent webhook via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhook.ID = nonexistentID

			err := testClients.main.UpdateWebhook(ctx, exampleWebhook)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be updateable via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhook, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, createdWebhook, err)

			// Change webhook.
			createdWebhook.Name = reverseString(createdWebhook.Name)
			exampleWebhook.Name = createdWebhook.Name

			assert.NoError(t, testClients.main.UpdateWebhook(ctx, createdWebhook))

			// Fetch webhook.
			actual, err := testClients.main.GetWebhook(ctx, createdWebhook.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := testClients.admin.GetAuditLogForWebhook(ctx, createdWebhook.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.WebhookCreationEvent},
				{EventType: audit.WebhookUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdWebhook.ID, audit.WebhookAssignmentKey)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, actual.ID))
		})
	}
}

func (s *TestSuite) TestWebhooksArchiving() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should fail to archive a non-existent webhook via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.main.ArchiveWebhook(ctx, nonexistentID))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be archived via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhook, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, createdWebhook, err)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForWebhook(ctx, createdWebhook.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.WebhookCreationEvent},
				{EventType: audit.WebhookArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdWebhook.ID, audit.WebhookAssignmentKey)
		})
	}
}

func (s *TestSuite) TestWebhooksAuditing() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should return an error when auditing a non-existent webhook via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhook.ID = nonexistentID

			// fetch audit log entries
			x, err := testClients.admin.GetAuditLogForWebhook(ctx, exampleWebhook.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should only be auditable to admins via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, premade, err)

			// Change webhook.
			premade.Name = reverseString(premade.Name)
			exampleWebhook.Name = premade.Name
			assert.NoError(t, testClients.main.UpdateWebhook(ctx, premade))

			// fetch audit log entries
			actual, err := testClients.main.GetAuditLogForWebhook(ctx, premade.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, premade.ID))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be auditable via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, premade, err)

			// Change webhook.
			premade.Name = reverseString(premade.Name)
			exampleWebhook.Name = premade.Name
			assert.NoError(t, testClients.main.UpdateWebhook(ctx, premade))

			// fetch audit log entries
			actual, err := testClients.admin.GetAuditLogForWebhook(ctx, premade.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 2)

			// Clean up item.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, premade.ID))
		})
	}
}
