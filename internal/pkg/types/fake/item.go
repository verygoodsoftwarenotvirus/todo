package fake

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeItem builds a faked item.
func BuildFakeItem() *types.Item {
	return &types.Item{
		ID:            fake.Uint64(),
		Name:          fake.Word(),
		Details:       fake.Word(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		BelongsToUser: fake.Uint64(),
	}
}

// BuildFakeItemList builds a faked ItemList.
func BuildFakeItemList() *types.ItemList {
	exampleItem1 := BuildFakeItem()
	exampleItem2 := BuildFakeItem()
	exampleItem3 := BuildFakeItem()

	return &types.ItemList{
		Pagination: types.Pagination{
			Page:  1,
			Limit: 20,
		},
		Items: []types.Item{
			*exampleItem1,
			*exampleItem2,
			*exampleItem3,
		},
	}
}

// BuildFakeItemUpdateInputFromItem builds a faked ItemUpdateInput from an item.
func BuildFakeItemUpdateInputFromItem(item *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:          item.Name,
		Details:       item.Details,
		BelongsToUser: item.BelongsToUser,
	}
}

// BuildFakeItemCreationInput builds a faked ItemCreationInput.
func BuildFakeItemCreationInput() *types.ItemCreationInput {
	item := BuildFakeItem()
	return BuildFakeItemCreationInputFromItem(item)
}

// BuildFakeItemCreationInputFromItem builds a faked ItemCreationInput from an item.
func BuildFakeItemCreationInputFromItem(item *types.Item) *types.ItemCreationInput {
	return &types.ItemCreationInput{
		Name:          item.Name,
		Details:       item.Details,
		BelongsToUser: item.BelongsToUser,
	}
}
