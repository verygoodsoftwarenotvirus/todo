package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

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

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func TestWebhooks(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be createable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, premade)

			// Clean up.
			err = testClient.ArchiveWebhook(ctx, premade.ID)
			assert.NoError(t, err)

			actual, err := testClient.GetWebhook(ctx, premade.ID)
			checkValueAndError(t, actual, err)
			checkWebhookEquality(t, exampleWebhook, actual)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Fetch webhook.
			_, err := testClient.GetWebhook(ctx, nonexistentID)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Fetch webhook.
			actual, err := testClient.GetWebhook(ctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, actual)

			// Clean up.
			err = testClient.ArchiveWebhook(ctx, actual.ID)
			assert.NoError(t, err)
		})

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhooks.
			var expected []*types.Webhook
			for i := 0; i < 5; i++ {
				exampleWebhook := fakes.BuildFakeWebhook()
				exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
				createdWebhook, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
				checkValueAndError(t, createdWebhook, err)

				expected = append(expected, createdWebhook)
			}

			// Assert webhook list equality.
			actual, err := testClient.GetWebhooks(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Webhooks))

			// Clean up.
			for _, webhook := range actual.Webhooks {
				err = testClient.ArchiveWebhook(ctx, webhook.ID)
				assert.NoError(t, err)
			}
		})

		subtest.Run("should only allow users to see their own webhooks", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, clientA := createUserAndClientForTest(ctx, t)
			_, clientB := createUserAndClientForTest(ctx, t)

			// create webhook for user A.
			wciA := fakes.BuildFakeWebhookCreationInput()
			webhookA, err := clientA.CreateWebhook(ctx, wciA)
			checkValueAndError(t, webhookA, err)

			// create webhook for user B.
			wciB := fakes.BuildFakeWebhookCreationInput()
			webhookB, err := clientB.CreateWebhook(ctx, wciB)
			checkValueAndError(t, webhookB, err)

			i, err := clientB.GetWebhook(ctx, webhookA.ID)
			assert.Nil(t, i)
			assert.Error(t, err, "should experience error trying to fetch entry they're not authorized for")

			// Clean up.
			assert.NoError(t, clientA.ArchiveWebhook(ctx, webhookA.ID))
			assert.NoError(t, clientB.ArchiveWebhook(ctx, webhookB.ID))
		})
	})

	test.Run("Updating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhook.ID = nonexistentID

			err := testClient.UpdateWebhook(ctx, exampleWebhook)
			assert.Error(t, err)
		})

		subtest.Run("it should be updatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Change webhook.
			premade.Name = reverse(premade.Name)
			exampleWebhook.Name = premade.Name
			err = testClient.UpdateWebhook(ctx, premade)
			assert.NoError(t, err)

			// Fetch webhook.
			actual, err := testClient.GetWebhook(ctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert webhook equality.
			checkWebhookEquality(t, exampleWebhook, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			// Clean up.
			err = testClient.ArchiveWebhook(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Clean up.
			err = testClient.ArchiveWebhook(ctx, premade.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhook.ID = nonexistentID

			x, err := adminClient.GetAuditLogForWebhook(ctx, exampleWebhook.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should be auditable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Change webhook.
			premade.Name = reverse(premade.Name)
			exampleWebhook.Name = premade.Name
			assert.NoError(t, testClient.UpdateWebhook(ctx, premade))

			// fetch audit log entries
			actual, err := adminClient.GetAuditLogForWebhook(ctx, premade.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 2)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveWebhook(ctx, premade.ID))
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create webhook.
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			premade, err := testClient.CreateWebhook(ctx, exampleWebhookInput)
			checkValueAndError(t, premade, err)

			// Change webhook.
			premade.Name = reverse(premade.Name)
			exampleWebhook.Name = premade.Name
			assert.NoError(t, testClient.UpdateWebhook(ctx, premade))

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForWebhook(ctx, premade.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveWebhook(ctx, premade.ID))
		})
	})
}
