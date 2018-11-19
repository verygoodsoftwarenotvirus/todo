package integration

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/bxcodec/faker"
	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Auth", func() {
		g.It("should reject an unauthenticated request")
		g.Describe("credentials", func() {
			g.It("should accept a valid cookie")
			g.It("should reject a valid cookie")
			g.It("should accept a valid auth key")
			g.It("should reject an invalid auth key")
		})
	})
}

func TestItems(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Items", func() {

		g.It("Should create an item", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			actual, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			assert.NoError(t, err)
			assert.NotNil(t, actual)

			// Assert item equality
			assert.NotZero(t, actual.ID)
			assert.Equal(t, expected.Name, actual.Name)
			assert.Equal(t, expected.Details, actual.Details)
			assert.NotZero(t, actual.CreatedOn)
			assert.Nil(t, actual.CompletedOn)

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)
		})

		g.It("Should return a pre-made item", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			assert.NoError(t, err)
			assert.NotNil(t, premade)

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			assert.NoError(t, err)

			// Assert item equality
			assert.NotZero(t, actual.ID)
			assert.Equal(t, expected.Name, actual.Name)
			assert.Equal(t, expected.Details, actual.Details)
			assert.NotZero(t, actual.CreatedOn)
			assert.Nil(t, actual.CompletedOn)

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)

		})

		g.It("Should error when fetching a nonexistent item", func() {
			// Fetch item
			actual, err := todoClient.GetItem(nonexistentID)
			assert.NoError(t, err)
			assert.Nil(t, actual)
		})

		g.It("Should return a list of pre-made items", func() {
			// Create items
			fake := faker.GetLorem()
			expected := []*models.Item{}
			for i := 0; i < 5; i++ {
				ii := &models.ItemInput{
					Name:    fake.Word(),
					Details: fake.Sentence(),
				}
				item, err := todoClient.CreateItem(ii)
				assert.NoError(t, err)
				expected = append(expected, item)
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(nil)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			require.True(t, len(expected) >= len(actual))

			// Clean up
			for _, item := range actual {
				assert.NoError(t, todoClient.DeleteItem(item.ID))
			}
		})

		g.It("Should update a item", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			assert.NoError(t, err)
			assert.NotNil(t, premade)

			// Change item
			expectedName, expectedDetails := "new name", "new details"
			premade.Name, premade.Details = expectedName, expectedDetails
			err = todoClient.UpdateItem(premade)
			assert.NoError(t, err)

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			assert.NoError(t, err)

			// Assert item equality
			assert.Equal(t, expected.Name, actual.Name)
			assert.Equal(t, expected.Details, actual.Details)
			assert.NotZero(t, actual.UpdatedOn)
			assert.Nil(t, actual.CompletedOn)

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)
		})

		g.It("Should delete a item", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			assert.NoError(t, err)
			assert.NotNil(t, premade)

			// Clean up
			err = todoClient.DeleteItem(premade.ID)
			assert.NoError(t, err)
		})
	})

}
