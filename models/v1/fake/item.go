package fakemodels

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeItem builds a faked item
func BuildFakeItem() *models.Item {
	return &models.Item{
		ID:            fake.Uint64(),
		Name:          fake.Word(),
		Details:       fake.Word(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		BelongsToUser: fake.Uint64(),
	}
}

// BuildFakeItemList builds a faked ItemList
func BuildFakeItemList() *models.ItemList {
	exampleItem1 := BuildFakeItem()
	exampleItem2 := BuildFakeItem()
	exampleItem3 := BuildFakeItem()

	return &models.ItemList{
		Pagination: models.Pagination{
			Page:       1,
			Limit:      20,
			TotalCount: 3,
		},
		Items: []models.Item{
			*exampleItem1,
			*exampleItem2,
			*exampleItem3,
		},
	}
}

// BuildFakeItemUpdateInputFromItem builds a faked ItemUpdateInput from an item
func BuildFakeItemUpdateInputFromItem(item *models.Item) *models.ItemUpdateInput {
	return &models.ItemUpdateInput{
		Name:          item.Name,
		Details:       item.Details,
		BelongsToUser: item.BelongsToUser,
	}
}

// BuildFakeItemCreationInput builds a faked ItemCreationInput
func BuildFakeItemCreationInput() *models.ItemCreationInput {
	item := BuildFakeItem()
	return BuildFakeItemCreationInputFromItem(item)
}

// BuildFakeItemCreationInputFromItem builds a faked ItemCreationInput from an item
func BuildFakeItemCreationInputFromItem(item *models.Item) *models.ItemCreationInput {
	return &models.ItemCreationInput{
		Name:          item.Name,
		Details:       item.Details,
		BelongsToUser: item.BelongsToUser,
	}
}
