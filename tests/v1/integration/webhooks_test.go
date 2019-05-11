package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkWebhookEquality(t *testing.T, expected, actual *models.Webhook) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ContentType, actual.ContentType)
	assert.Equal(t, expected.URL, actual.URL)
	assert.Equal(t, expected.Method, actual.Method)
	assert.NotZero(t, actual.CreatedOn)
}

func buildDummyWebhookInput() *models.WebhookInput {
	x := &models.WebhookInput{
		Name:        fake.Word(),
		URL:         fake.DomainName(),
		ContentType: "application/json",
		Method:      "POST",
	}
	return x
}

func buildDummyWebhook(t *testing.T) *models.Webhook {
	t.Helper()

	y, err := todoClient.CreateWebhook(context.Background(), buildDummyWebhookInput())
	require.NoError(t, err)
	return y
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

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be createable", func(t *testing.T) {
			tctx := context.Background()

			// Create webhook
			input := buildDummyWebhookInput()
			expected := &models.Webhook{
				Name:        input.Name,
				URL:         input.URL,
				ContentType: input.ContentType,
				Method:      input.Method,
			}
			premade, err := todoClient.CreateWebhook(
				tctx,
				&models.WebhookInput{
					Name:        expected.Name,
					ContentType: expected.ContentType,
					URL:         expected.URL,
					Method:      expected.Method,
				})
			checkValueAndError(t, premade, err)

			// Assert webhook equality
			checkWebhookEquality(t, expected, premade)

			// Clean up
			err = todoClient.DeleteWebhook(tctx, premade.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetWebhook(tctx, premade.ID)
			checkValueAndError(t, actual, err)
			checkWebhookEquality(t, expected, actual)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := context.Background()

			// Create webhooks
			var expected []*models.Webhook
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyWebhook(t))
			}

			// Assert webhook list equality
			actual, err := todoClient.GetWebhooks(tctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Webhooks))

			// Clean up
			for _, webhook := range actual.Webhooks {
				err = todoClient.DeleteWebhook(tctx, webhook.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()

			// Fetch webhook
			_, err := todoClient.GetWebhook(tctx, nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := context.Background()

			// Create webhook
			input := buildDummyWebhookInput()
			expected := &models.Webhook{
				Name:        input.Name,
				URL:         input.URL,
				ContentType: input.ContentType,
				Method:      input.Method,
			}
			premade, err := todoClient.CreateWebhook(tctx, &models.WebhookInput{
				Name:        expected.Name,
				ContentType: expected.ContentType,
				URL:         expected.URL,
				Method:      expected.Method,
			})
			checkValueAndError(t, premade, err)

			// Fetch webhook
			actual, err := todoClient.GetWebhook(tctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert webhook equality
			checkWebhookEquality(t, expected, actual)

			// Clean up
			err = todoClient.DeleteWebhook(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should return an error when trying to update something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()

			err := todoClient.UpdateWebhook(tctx, &models.Webhook{ID: nonexistentID})
			assert.Error(t, err)

		})

		T.Run("it should be updatable", func(t *testing.T) {
			tctx := context.Background()

			// Create webhook
			input := buildDummyWebhookInput()
			expected := &models.Webhook{
				Name:        input.Name,
				URL:         input.URL,
				ContentType: input.ContentType,
				Method:      input.Method,
			}
			premade, err := todoClient.CreateWebhook(
				tctx,
				&models.WebhookInput{
					Name:        expected.Name,
					ContentType: expected.ContentType,
					URL:         expected.URL,
					Method:      expected.Method,
				},
			)
			checkValueAndError(t, premade, err)

			// Change webhook
			premade.Name = reverse(premade.Name)
			expected.Name = premade.Name
			err = todoClient.UpdateWebhook(tctx, premade)
			assert.NoError(t, err)

			// Fetch webhook
			actual, err := todoClient.GetWebhook(tctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert webhook equality
			checkWebhookEquality(t, expected, actual)
			assert.NotNil(t, actual.UpdatedOn)

			// Clean up
			err = todoClient.DeleteWebhook(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := context.Background()

			// Create webhook
			input := buildDummyWebhookInput()
			expected := &models.Webhook{
				Name:        input.Name,
				URL:         input.URL,
				ContentType: input.ContentType,
				Method:      input.Method,
			}
			premade, err := todoClient.CreateWebhook(tctx, &models.WebhookInput{
				Name:        expected.Name,
				ContentType: expected.ContentType,
				URL:         expected.URL,
				Method:      expected.Method,
			})
			checkValueAndError(t, premade, err)

			// Clean up
			err = todoClient.DeleteWebhook(tctx, premade.ID)
			assert.NoError(t, err)
		})
	})
}
