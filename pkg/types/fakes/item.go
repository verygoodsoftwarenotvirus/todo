package fakes

import (
	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFakeItem builds a faked item.
func BuildFakeItem() *types.Item {
	return &types.Item{
		ID:               ksuid.New().String(),
		Name:             fake.Word(),
		Details:          fake.Word(),
		CreatedOn:        uint64(uint32(fake.Date().Unix())),
		BelongsToAccount: fake.UUID(),
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

// BuildFakeItemUpdateRequestInput builds a faked ItemUpdateRequestInput from an item.
func BuildFakeItemUpdateRequestInput() *types.ItemUpdateRequestInput {
	item := BuildFakeItem()
	return &types.ItemUpdateRequestInput{
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
	}
}

// BuildFakeItemUpdateRequestInputFromItem builds a faked ItemUpdateRequestInput from an item.
func BuildFakeItemUpdateRequestInputFromItem(item *types.Item) *types.ItemUpdateRequestInput {
	return &types.ItemUpdateRequestInput{
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
	}
}

// BuildFakeItemCreationRequestInput builds a faked ItemCreationRequestInput.
func BuildFakeItemCreationRequestInput() *types.ItemCreationRequestInput {
	item := BuildFakeItem()
	return BuildFakeItemCreationRequestInputFromItem(item)
}

// BuildFakeItemCreationRequestInputFromItem builds a faked ItemCreationRequestInput from an item.
func BuildFakeItemCreationRequestInputFromItem(item *types.Item) *types.ItemCreationRequestInput {
	return &types.ItemCreationRequestInput{
		ID:               item.ID,
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
	}
}

// BuildFakeItemDatabaseCreationInput builds a faked ItemDatabaseCreationInput.
func BuildFakeItemDatabaseCreationInput() *types.ItemDatabaseCreationInput {
	item := BuildFakeItem()
	return BuildFakeItemDatabaseCreationInputFromItem(item)
}

// BuildFakeItemDatabaseCreationInputFromItem builds a faked ItemDatabaseCreationInput from an item.
func BuildFakeItemDatabaseCreationInputFromItem(item *types.Item) *types.ItemDatabaseCreationInput {
	return &types.ItemDatabaseCreationInput{
		ID:               item.ID,
		Name:             item.Name,
		Details:          item.Details,
		BelongsToAccount: item.BelongsToAccount,
	}
}
