package integration

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
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

func (s *TestSuite) TestWebhooks_Creating() {
	s.runForEachClientExcept("should be createable", func(testClients *testClientWrapper) func() {
		return func() {
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

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))
		}
	})
}

func (s *TestSuite) TestWebhooks_Reading_Returns404ForNonexistentWebhook() {
	s.runForEachClientExcept("should fail to read non-existent webhook", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Fetch webhook.
			_, err := testClients.main.GetWebhook(ctx, nonexistentID)
			assert.Error(t, err)
		}
	})
}

func (s *TestSuite) TestWebhooks_Reading() {
	s.runForEachClientExcept("should be able to be read", func(testClients *testClientWrapper) func() {
		return func() {
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
		}
	})
}

func (s *TestSuite) TestWebhooks_Listing() {
	s.runForEachClientExcept("should be able to be read in a list", func(testClients *testClientWrapper) func() {
		return func() {
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
		}
	})
}

func (s *TestSuite) TestWebhooks_Archiving_Returns404ForNonexistentWebhook() {
	s.runForEachClientExcept("should fail to archive a non-existent webhook", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.main.ArchiveWebhook(ctx, nonexistentID))
		}
	})
}

func (s *TestSuite) TestWebhooks_Archiving() {
	s.runForEachClientExcept("should be able to be archived", func(testClients *testClientWrapper) func() {
		return func() {
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
		}
	})
}
