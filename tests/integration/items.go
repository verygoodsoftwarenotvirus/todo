package integration

import (
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/bxcodec/faker"
)

type ItemManager struct {
	*client.V1Client
	items []*models.Item
}

func NewItemManager(c *client.V1Client) *ItemManager {
	return &ItemManager{
		V1Client: c,
		items:    []*models.Item{},
	}
}

func (im *ItemManager) Add(count int) []*models.Item {
	fake := faker.GetLorem()
	items := []*models.Item{}
	for i := 0; i < count; i++ {
		ii := &models.ItemInput{
			Name:    fake.Word(),
			Details: fake.Sentence(),
		}
		item, err := im.CreateItem(ii)
		if err != nil {
			panic(err)
		}
		items = append(items, item)
	}
	im.items = append(im.items, items...)
	return items
}

func (im *ItemManager) CleanUp() {
	for _, item := range im.items {
		if err := im.DeleteItem(item.ID); err != nil {
			log.Printf("error deleting item %d: %v", item.ID, err)
		}
	}
}
