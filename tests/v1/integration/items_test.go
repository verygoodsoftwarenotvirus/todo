package integration

import (
	"context"
	"fmt"
	"testing"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func checkItemEquality(t *testing.T, expected, actual *models.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for ID %d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for ID %d to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func TestItems(test *testing.T) {
	test.Run("Creating", func(t *testing.T) {
		t.Run("should be createable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, createdItem)

			// Clean up.
			err = todoClient.ArchiveItem(ctx, createdItem.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)
			checkItemEquality(t, exampleItem, actual)
			assert.NotNil(t, actual.ArchivedOn)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Listing", func(t *testing.T) {
		t.Run("should be able to be read in a list", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create items.
			var expected []*models.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItem := fakemodels.BuildFakeItem()
				exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
				createdItem, itemCreationErr := todoClient.CreateItem(ctx, exampleItemInput)
				checkValueAndError(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			// Assert item list equality.
			actual, err := todoClient.GetItems(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			// Clean up.
			for _, createdItem := range actual.Items {
				err = todoClient.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Searching", func(t *testing.T) {
		t.Run("should be able to be search for items", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create items.
			exampleItem := fakemodels.BuildFakeItem()
			var expected []*models.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)
				createdItem, itemCreationErr := todoClient.CreateItem(ctx, exampleItemInput)
				checkValueAndError(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// Assert item list equality.
			actual, err := todoClient.SearchItems(ctx, exampleItem.Name, exampleLimit)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
			for _, createdItem := range expected {
				err = todoClient.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})

		t.Run("should only receive your own items", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// create user and oauth2 client A.
			userA, err := testutil.CreateObligatoryUser(urlToUse, debug)
			require.NoError(t, err)

			ca, err := testutil.CreateObligatoryClient(ctx, urlToUse, userA)
			require.NoError(t, err)

			clientA, err := client.NewClient(
				ctx,
				ca.ClientID,
				ca.ClientSecret,
				todoClient.URL,
				noop.NewLogger(),
				buildHTTPClient(),
				ca.Scopes,
				true,
			)
			checkValueAndError(test, clientA, err)

			// Create items for user A.
			exampleItemA := fakemodels.BuildFakeItem()
			var createdForA []*models.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInputA := fakemodels.BuildFakeItemCreationInputFromItem(exampleItemA)
				exampleItemInputA.Name = fmt.Sprintf("%s %d", exampleItemInputA.Name, i)

				createdItem, itemCreationErr := clientA.CreateItem(ctx, exampleItemInputA)
				checkValueAndError(t, createdItem, itemCreationErr)

				createdForA = append(createdForA, createdItem)
			}

			exampleLimit := uint8(20)
			query := exampleItemA.Name

			// create user and oauth2 client B.
			userB, err := testutil.CreateObligatoryUser(urlToUse, debug)
			require.NoError(t, err)

			cb, err := testutil.CreateObligatoryClient(ctx, urlToUse, userB)
			require.NoError(t, err)

			clientB, err := client.NewClient(
				ctx,
				cb.ClientID,
				cb.ClientSecret,
				todoClient.URL,
				noop.NewLogger(),
				buildHTTPClient(),
				cb.Scopes,
				true,
			)
			checkValueAndError(test, clientB, err)

			// Create items for user B.
			exampleItemB := fakemodels.BuildFakeItem()
			exampleItemB.Name = reverse(exampleItemA.Name)
			var createdForB []*models.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInputB := fakemodels.BuildFakeItemCreationInputFromItem(exampleItemB)
				exampleItemInputB.Name = fmt.Sprintf("%s %d", exampleItemInputB.Name, i)

				createdItem, itemCreationErr := clientB.CreateItem(ctx, exampleItemInputB)
				checkValueAndError(t, createdItem, itemCreationErr)

				createdForB = append(createdForB, createdItem)
			}

			expected := createdForA

			// Assert item list equality.
			actual, err := clientA.SearchItems(ctx, query, exampleLimit)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
			for _, createdItem := range createdForA {
				err = clientA.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}

			for _, createdItem := range createdForB {
				err = clientB.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("ExistenceChecking", func(t *testing.T) {
		t.Run("it should return false with no error when checking something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Attempt to fetch nonexistent item.
			actual, err := todoClient.ItemExists(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.False(t, actual)
		})

		t.Run("it should return true with no error when the relevant item exists", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item.
			actual, err := todoClient.ItemExists(ctx, createdItem.ID)
			assert.NoError(t, err)
			assert.True(t, actual)

			// Clean up item.
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Reading", func(t *testing.T) {
		t.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Attempt to fetch nonexistent item.
			_, err := todoClient.GetItem(ctx, nonexistentID)
			assert.Error(t, err)
		})

		t.Run("it should be readable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item.
			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, actual)

			// Clean up item.
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Updating", func(t *testing.T) {
		t.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			exampleItem := fakemodels.BuildFakeItem()
			exampleItem.ID = nonexistentID

			assert.Error(t, todoClient.UpdateItem(ctx, exampleItem))
		})

		t.Run("it should be updatable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			updateTo := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Change item.
			assert.NotEmpty(t, createdItem.Update(updateTo.ToUpdateInput()))
			err = todoClient.UpdateItem(ctx, createdItem)
			assert.NoError(t, err)

			// Fetch item.
			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, updateTo, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			// Clean up item.
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Auditing", func(t *testing.T) {
		t.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			exampleItem := fakemodels.BuildFakeItem()
			exampleItem.ID = nonexistentID

			x, err := adminClient.GetAuditLogForItem(ctx, exampleItem.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		t.Run("it should be auditable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			updateTo := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Change item.
			expectedChanges := createdItem.Update(updateTo.ToUpdateInput())
			require.NotEmpty(t, expectedChanges)
			err = todoClient.UpdateItem(ctx, createdItem)
			assert.NoError(t, err)

			// Fetch item.
			updated, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, updated, err)

			// Assert item equality.
			checkItemEquality(t, updateTo, updated)
			assert.NotNil(t, updated.LastUpdatedOn)

			// fetch audit log entries
			actual, err := adminClient.GetAuditLogForItem(ctx, updated.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 2)

			// Clean up item.
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Deleting", func(t *testing.T) {
		t.Run("it should return an error when trying to delete something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			assert.Error(t, todoClient.ArchiveItem(ctx, nonexistentID))
		})

		t.Run("should be able to be deleted", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakemodels.BuildFakeItem()
			exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Clean up item.
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})
}
