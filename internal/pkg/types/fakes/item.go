package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeItem builds a faked item.
func BuildFakeItem() *types.Item {
	return &types.Item{
		ID:               uint64(fake.Uint32()),
		ExternalID:       fake.UUID(),
		Name:             fake.Word(),
		Details:          fake.Word(),
		CreatedOn:        uint64(uint32(fake.Date().Unix())),
		BelongsToAccount: fake.Uint64(),
	}
}

// BuildFakeItemList builds a faked ItemList.
func BuildFakeItemList() *types.ItemList {
	var examples []*types.Item
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeItem())
	}

	return &types.ItemList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Items: examples,
	}
}

// BuildFakeItemUpdateInputFromItem builds a faked ItemUpdateInput from an item.
func BuildFakeItemUpdateInputFromItem(item *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
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
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
	}
}
