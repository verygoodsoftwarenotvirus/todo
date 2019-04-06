package todoproto

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// ToModelQueryFilter attaches to QueryFilter to provide itself a way to convert to our models' package version of a QueryFilter
func (qf *QueryFilter) ToModelQueryFilter() *models.QueryFilter {
	var sb = models.SortAscending
	if qf.SortBy == string(models.SortDescending) {
		sb = models.SortAscending
	}

	return &models.QueryFilter{
		Page:          qf.Page,
		Limit:         qf.Limit,
		CreatedAfter:  qf.CreatedAfter,
		CreatedBefore: qf.CreatedBefore,
		UpdatedAfter:  qf.UpdatedAfter,
		UpdatedBefore: qf.UpdatedBefore,
		IncludeAll:    qf.IncludeAll,
		SortBy:        sb,
	}
}

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
