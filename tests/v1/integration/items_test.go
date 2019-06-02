package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkItemEquality(t *testing.T, expected, actual *models.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func buildDummyItem(t *testing.T) *models.Item {
	t.Helper()

	x := &models.ItemInput{
		Name:    fake.Word(),
		Details: fake.Sentence(),
	}
	y, err := todoClient.CreateItem(context.Background(), x)
	require.NoError(t, err)
	return y
}

func TestItems(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be createable", func(t *testing.T) {
			tctx := context.Background()

			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(
				tctx,
				&models.ItemInput{
					Name:    expected.Name,
					Details: expected.Details,
				})
			checkValueAndError(t, premade, err)

			// Assert item equality
			checkItemEquality(t, expected, premade)

			// Clean up
			err = todoClient.DeleteItem(tctx, premade.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetItem(tctx, premade.ID)
			checkValueAndError(t, actual, err)
			checkItemEquality(t, expected, actual)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := context.Background()

			// Create items
			var expected []*models.Item
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(tctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Items))

			// Clean up
			for _, item := range actual.Items {
				err = todoClient.DeleteItem(tctx, item.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()

			// Fetch item
			_, err := todoClient.GetItem(tctx, nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := context.Background()

			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(tctx, &models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Fetch item
			actual, err := todoClient.GetItem(tctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)

			// Clean up
			err = todoClient.DeleteItem(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should return an error when trying to update something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()

			err := todoClient.UpdateItem(tctx, &models.Item{ID: nonexistentID})
			assert.Error(t, err)

		})

		T.Run("it should be updatable", func(t *testing.T) {
			tctx := context.Background()

			// Create item
			expected := &models.Item{Name: "new name", Details: "new details"}
			premade, err := todoClient.CreateItem(
				tctx,
				&models.ItemInput{
					Name:    "old name",
					Details: "old details",
				},
			)
			checkValueAndError(t, premade, err)

			// Change item
			premade.Name, premade.Details = expected.Name, expected.Details
			err = todoClient.UpdateItem(tctx, premade)
			assert.NoError(t, err)

			// Fetch item
			actual, err := todoClient.GetItem(tctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)
			assert.NotNil(t, actual.UpdatedOn)

			// Clean up
			err = todoClient.DeleteItem(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := context.Background()

			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(tctx, &models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Clean up
			err = todoClient.DeleteItem(tctx, premade.ID)
			assert.NoError(t, err)
		})
	})
}
