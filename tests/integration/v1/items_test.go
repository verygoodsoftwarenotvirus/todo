package integration

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
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
		Name:    faker.Lorem{}.Word(),
		Details: faker.Lorem{}.Sentence(),
	}
	y, err := todoClient.CreateItem(x)
	assert.NoError(t, err)
	return y
}

func TestItems(test *testing.T) {
	// test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be createable", func(t *testing.T) {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Assert item equality
			checkItemEquality(t, expected, premade)

			// Clean up
			err = todoClient.DeleteItem(premade.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetItem(premade.ID)
			checkValueAndError(t, actual, err)
			checkItemEquality(t, expected, actual)
			assert.NotZero(t, actual.CompletedOn)
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			// Fetch item
			_, err := todoClient.GetItem(nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should be updatable", func(t *testing.T) {
			// Create item
			expected := &models.Item{Name: "new name", Details: "new details"}
			premade, err := todoClient.CreateItem(
				&models.ItemInput{
					Name:    "old name",
					Details: "old details",
				},
			)
			checkValueAndError(t, premade, err)

			// Change item
			premade.Name, premade.Details = expected.Name, expected.Details
			err = todoClient.UpdateItem(premade)
			assert.NoError(t, err)

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			assert.NoError(t, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)
			assert.NotNil(t, actual.UpdatedOn)

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Clean up
			err = todoClient.DeleteItem(premade.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			// Create items
			expected := []*models.Item{}
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Items))

			// Clean up
			for _, item := range actual.Items {
				err := todoClient.DeleteItem(item.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Counting", func(T *testing.T) {
		T.Run("it should be able to be counted", func(t *testing.T) {
			t.Skip()
		})
	})
}
