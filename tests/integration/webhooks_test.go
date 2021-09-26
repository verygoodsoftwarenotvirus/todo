package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
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
	s.runForCookieClient("should be creatable and readable and deletable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			stopChan := make(chan bool, 1)
			notificationsChan, err := testClients.main.SubscribeToDataChangeNotifications(ctx, stopChan)
			require.NotNil(t, notificationsChan)
			require.NoError(t, err)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhookID, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			require.NoError(t, err)

			n := <-notificationsChan
			assert.Equal(t, n.DataType, types.WebhookDataType)
			require.NotNil(t, n.Webhook)
			checkWebhookEquality(t, exampleWebhook, n.Webhook)

			createdWebhook, err := testClients.main.GetWebhook(ctx, createdWebhookID)
			requireNotNilAndNoProblems(t, createdWebhook, err)

			// assert webhook equality
			checkWebhookEquality(t, exampleWebhook, createdWebhook)

			actual, err := testClients.main.GetWebhook(ctx, createdWebhook.ID)
			requireNotNilAndNoProblems(t, actual, err)
			checkWebhookEquality(t, exampleWebhook, actual)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))
		}
	})

	s.runForPASETOClient("should be creatable and readable and deletable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhookID, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			require.NoError(t, err)

			var createdWebhook *types.Webhook
			checkFunc := func() bool {
				createdWebhook, err = testClients.main.GetWebhook(ctx, createdWebhookID)
				return assert.NotNil(t, createdWebhook) && assert.NoError(t, err)
			}
			assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)

			// assert webhook equality
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

func (s *TestSuite) TestWebhooks_Listing() {
	s.runForCookieClient("should be able to be read in a list", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			stopChan := make(chan bool, 1)
			notificationsChan, err := testClients.main.SubscribeToDataChangeNotifications(ctx, stopChan)
			require.NotNil(t, notificationsChan)
			require.NoError(t, err)

			// Create webhooks.
			var expected []*types.Webhook
			for i := 0; i < 5; i++ {
				// Create webhook.
				exampleWebhook := fakes.BuildFakeWebhook()
				exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
				createdWebhookID, webhookCreationErr := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
				require.NoError(t, webhookCreationErr)

				n := <-notificationsChan
				assert.Equal(t, n.DataType, types.WebhookDataType)
				require.NotNil(t, n.Webhook)
				checkWebhookEquality(t, exampleWebhook, n.Webhook)

				createdWebhook, webhookCreationErr := testClients.main.GetWebhook(ctx, createdWebhookID)
				requireNotNilAndNoProblems(t, createdWebhook, webhookCreationErr)

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

	s.runForPASETOClient("should be able to be read in a list", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create webhooks.
			var expected []*types.Webhook
			for i := 0; i < 5; i++ {
				// Create webhook.
				exampleWebhook := fakes.BuildFakeWebhook()
				exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
				createdWebhookID, err := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
				require.NoError(t, err)

				var createdWebhook *types.Webhook
				checkFunc := func() bool {
					createdWebhook, err = testClients.main.GetWebhook(ctx, createdWebhookID)
					return assert.NotNil(t, createdWebhook) && assert.NoError(t, err)
				}
				assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)

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
