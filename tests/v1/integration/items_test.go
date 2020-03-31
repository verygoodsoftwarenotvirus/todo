package integration

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
)

func checkItemEquality(t *testing.T, expected, actual *models.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for ID %d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for ID %d to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func buildDummyItem(t *testing.T) *models.Item {
	t.Helper()

	ctx := context.Background()
	exampleItem := fakemodels.BuildFakeItem()
	exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
	y, err := todoClient.CreateItem(ctx, exampleInput)
	require.NoError(t, err)

	return y
}

func TestItems(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be createable", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create item
			exampleItem := fakemodels.BuildFakeItem()
			exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleInput)
			checkValueAndError(t, createdItem, err)

			// Assert item equality
			checkItemEquality(t, exampleItem, createdItem)

			// Clean up
			err = todoClient.ArchiveItem(ctx, createdItem.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)
			checkItemEquality(t, exampleItem, actual)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create items
			var expected []*models.Item
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			// Clean up
			for _, createdItem := range actual.Items {
				err = todoClient.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("ExistenceChecking", func(T *testing.T) {
		T.Run("it should return an error when trying to check something that does not exist", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Attempt to fetch nonexistent item
			actual, err := todoClient.ItemExists(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.False(t, actual)
		})

		T.Run("it should return 200 when the relevant item exists", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create item
			exampleItem := fakemodels.BuildFakeItem()
			exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item
			actual, err := todoClient.ItemExists(ctx, createdItem.ID)
			assert.NoError(t, err)
			assert.True(t, actual)

			// Clean up item
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Attempt to fetch nonexistent item
			_, err := todoClient.GetItem(ctx, nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create item
			exampleItem := fakemodels.BuildFakeItem()
			exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item
			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, exampleItem, actual)

			// Clean up item
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			exampleItem := fakemodels.BuildFakeItem()
			exampleItem.ID = nonexistentID

			assert.Error(t, todoClient.UpdateItem(ctx, exampleItem))
		})

		T.Run("it should be updatable", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create item
			exampleItem := fakemodels.BuildFakeItem()
			exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleInput)
			checkValueAndError(t, createdItem, err)

			// Change item
			createdItem.Update(exampleItem.ToUpdateInput())
			err = todoClient.UpdateItem(ctx, createdItem)
			assert.NoError(t, err)

			// Fetch item
			actual, err := todoClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, exampleItem, actual)
			assert.NotNil(t, actual.UpdatedOn)

			// Clean up item
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := trace.StartSpan(ctx, t.Name())
			defer span.End()

			// Create item
			exampleItem := fakemodels.BuildFakeItem()
			exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := todoClient.CreateItem(ctx, exampleInput)
			checkValueAndError(t, createdItem, err)

			// Clean up item
			assert.NoError(t, todoClient.ArchiveItem(ctx, createdItem.ID))
		})
	})
}
