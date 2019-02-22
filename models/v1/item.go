package models

import (
	"context"
)

// ItemHandler describes a structure capable of storing items permanently
type ItemHandler interface {
	GetItem(ctx context.Context, itemID, userID uint64) (*Item, error)
	GetItemCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
	GetItems(ctx context.Context, filter *QueryFilter, userID uint64) (*ItemList, error)
	CreateItem(ctx context.Context, input *ItemInput) (*Item, error)
	UpdateItem(ctx context.Context, updated *Item) error
	DeleteItem(ctx context.Context, id uint64, userID uint64) error
}

// Item represents an item
type Item struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Details     string  `json:"details"`
	CreatedOn   uint64  `json:"created_on"`
	UpdatedOn   *uint64 `json:"updated_on"`
	CompletedOn *uint64 `json:"completed_on"`
	BelongsTo   uint64  `json:"belongs_to"`
}

// Update merges an ItemInput with an Item
func (i *Item) Update(input *ItemInput) {
	if input.Name != "" || input.Name != i.Name {
		i.Name = input.Name
	}

	if input.Details != "" || input.Details != i.Details {
		i.Details = input.Details
	}
}

// ItemList represents a list of items
type ItemList struct {
	Pagination
	Items []Item `json:"items"`
}

// ItemInput represents what a user could set as input for items
type ItemInput struct {
	Name      string `json:"name"`
	Details   string `json:"details"`
	BelongsTo uint64 `json:"-"`
}
