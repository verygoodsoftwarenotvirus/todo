package integration

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

func buildDummyItem(t *testing.T) *models.Item {
	t.Helper()

	ii := &models.ItemInput{
		Name:    fake.Word(),
		Details: fake.Sentence(),
	}
	item, err := todoClient.CreateItem(ii)
	assert.NoError(t, err)
	return item
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
			assert.NotZero(t, actual.ID, "ID should be non-zero")
			assert.Equal(t, expected.Name, actual.Name, "expected .Name to be %q, but it was %q", expected.Name, actual.Name)
			assert.Equal(t, expected.Details, actual.Details, "expected .Details to be %q, but it was %q", expected.Details, actual.Details)
			assert.NotZero(t, actual.CreatedOn, ".CreatedOn should not be zero")
			assert.Nil(t, actual.CompletedOn, ".CompletedOn should be nil")
			assert.Nil(t, actual.UpdatedOn, ".UpdatedOn should be nil")

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
			assert.NotZero(t, actual.ID, "ID should be non-zero")
			assert.Equal(t, expected.Name, actual.Name, "expected .Name to be %q, but it was %q", expected.Name, actual.Name)
			assert.Equal(t, expected.Details, actual.Details, "expected .Details to be %q, but it was %q", expected.Details, actual.Details)
			assert.NotZero(t, actual.CreatedOn, ".CreatedOn should not be zero")
			assert.Nil(t, actual.CompletedOn, ".CompletedOn should be nil")
			assert.Nil(t, actual.UpdatedOn, ".UpdatedOn should be nil")

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			assert.NoError(t, err)
		})

		g.It("Should error when fetching a nonexistent item", func() {
			// Fetch item
			_, err := todoClient.GetItem(nonexistentID)
			assert.Error(t, err)
			// assert.Nil(t, actual)
		})

		g.It("Should return a list of pre-made items", func() {
			// Create items
			expected := []*models.Item{}
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(nil)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.True(t, len(expected) <= len(actual), "expected %d to be <= %d", len(expected), len(actual))

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
			assert.Equal(t, expectedName, actual.Name, "expected .Name to be %q, but it was %q", expected.Name, actual.Name)
			assert.Equal(t, expectedDetails, actual.Details, "expected .Details to be %q, but it was %q", expected.Details, actual.Details)
			assert.NotZero(t, actual.CreatedOn, ".CreatedOn should not be zero")
			assert.NotNil(t, actual.UpdatedOn, ".UpdatedOn should not be nil")
			assert.Nil(t, actual.CompletedOn, ".CompletedOn should be nil")

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

		// g.It("should serve a websocket feed", func() {
		// 	createdItems := []*models.Item{}
		// 	reportedItemsCount := len(createdItems)
		// 	feed, err := todoClient.ItemsFeed()
		// 	if err != nil {
		// 		g.Fail(err)
		// 		return
		// 	}

		// 	timeLimit := time.NewTimer(2 * time.Second)
		// 	ticker := time.NewTicker(250 * time.Millisecond)
		// 	defer func() {
		// 		ticker.Stop()
		// 	}()

		// 	done := false
		// 	for done {
		// 		select {
		// 		case <-feed:
		// 			reportedItemsCount += 1
		// 			t.Logf("item #%d came into feed. item count: %d", reportedItemsCount, len(createdItems))
		// 		case <-ticker.C:
		// 			t.Log("creating new item")
		// 			createdItems = append(createdItems, buildDummyItem(t))
		// 		case <-timeLimit.C:
		// 			t.Log("timer has gone off")
		// 			ticker.Stop()
		// 			assert.Equal(
		// 				t,
		// 				len(createdItems),
		// 				reportedItemsCount,
		// 				"expected number of created items (%d) to match the number of items that came through the websocket (%d)",
		// 				len(createdItems),
		// 				reportedItemsCount,
		// 			)
		// 			done = true
		// 		}
		// 	}
		// })
	})

}
