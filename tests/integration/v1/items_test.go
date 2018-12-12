package integration

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func buildDummyItem(t *testing.T) *models.Item {
	t.Helper()

	ii := &models.ItemInput{
		Name:    faker.Lorem{}.Word(),
		Details: faker.Lorem{}.Sentence(),
	}
	item, err := todoClient.CreateItem(ii)
	Ω(err).ShouldNot(HaveOccurred())
	return item
}

func TestItems(t *testing.T) {
	g := goblin.Goblin(t)

	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Items", func() {

		g.It("should be able to be created", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			actual, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())

			// Assert item equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(expected.Name).Should(Equal(actual.Name))
			Ω(expected.Details).Should(Equal(actual.Details))
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.CompletedOn).Should(BeNil())
			Ω(actual.UpdatedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should be able to be read", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(premade).ShouldNot(BeNil())

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			Ω(err).ShouldNot(HaveOccurred())

			// Assert item equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(expected.Name).Should(Equal(actual.Name))
			Ω(expected.Details).Should(Equal(actual.Details))
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.CompletedOn).Should(BeNil())
			Ω(actual.UpdatedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should return an error if it does not exist", func() {
			// Fetch item
			_, err := todoClient.GetItem(nonexistentID)
			Ω(err).Should(HaveOccurred())
			// assert.Nil(t, actual)
		})

		g.It("should be able to be read in a list", func() {
			// Create items
			expected := []*models.Item{}
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(nil)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())
			Ω(len(expected) <= len(actual.Items)).Should(BeTrue())

			// Clean up
			for _, item := range actual.Items {
				err := todoClient.DeleteItem(item.ID)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})

		g.It("should be able to be updated", func() {
			// Create item
			expected := &models.Item{Name: "new name", Details: "new details"}
			premade, err := todoClient.CreateItem(
				&models.ItemInput{
					Name:    "old name",
					Details: "old details",
				},
			)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(premade).ShouldNot(BeNil())

			// Change item
			premade.Name, premade.Details = expected.Name, expected.Details
			err = todoClient.UpdateItem(premade)
			Ω(err).ShouldNot(HaveOccurred())

			// Fetch item
			actual, err := todoClient.GetItem(premade.ID)
			Ω(err).ShouldNot(HaveOccurred())

			// Assert item equality
			Ω(actual.ID).Should(Equal(premade.ID))
			Ω(actual.Name).Should(Equal(expected.Name))
			Ω(actual.Details).Should(Equal(expected.Details))
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.UpdatedOn).ShouldNot(BeNil())
			Ω(actual.CompletedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteItem(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should be able to be deleted", func() {
			// Create item
			expected := &models.Item{Name: "name", Details: "details"}
			premade, err := todoClient.CreateItem(&models.ItemInput{
				Name: expected.Name, Details: expected.Details,
			})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(premade).ShouldNot(BeNil())

			// Clean up
			err = todoClient.DeleteItem(premade.ID)
			Ω(err).ShouldNot(HaveOccurred())
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
