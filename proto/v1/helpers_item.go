package todoproto

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// ProtoItemFromModel converts a models item into a gRPC Item.
func ProtoItemFromModel(i *models.Item) *Item {
	return &Item{
		Id:          i.ID,
		Name:        i.Name,
		Details:     i.Details,
		CreatedOn:   i.CreatedOn,
		UpdatedOn:   *i.UpdatedOn,
		CompletedOn: *i.CompletedOn,
		BelongsTo:   i.BelongsTo,
	}
}

// ProtoItemsFromModels converts a slice of models items into a gRPC Item slice.
func ProtoItemsFromModels(in []models.Item) (out []*Item) {
	for _, i := range in {
		out = append(out, ProtoItemFromModel(&i))
	}
	return
}

// ToModelItem converts a gRPC Item into a models item
func (i *Item) ToModelItem() *models.Item {
	return &models.Item{
		ID:          i.Id,
		Name:        i.Name,
		Details:     i.Details,
		CreatedOn:   i.CreatedOn,
		UpdatedOn:   &i.UpdatedOn,
		CompletedOn: &i.CompletedOn,
		BelongsTo:   i.BelongsTo,
	}
}

// ToItemInput converts a gRPC CreateItemRequest into a models ItemInput
func (r *CreateItemRequest) ToItemInput() *models.ItemInput {
	return &models.ItemInput{
		Name:    r.GetName(),
		Details: r.GetDetails(),
	}
}
