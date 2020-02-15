package models

import (
	"context"
	"net/http"
)

type (
	// Item represents an item
	Item struct {
		ID            uint64  `json:"id"`
		Name          string  `json:"name"`
		Details       string  `json:"details"`
		CreatedOn     uint64  `json:"created_on"`
		UpdatedOn     *uint64 `json:"updated_on"`
		ArchivedOn    *uint64 `json:"archived_on"`
		BelongsToUser uint64  `json:"belongs_to_user"`
	}

	// ItemList represents a list of items
	ItemList struct {
		Pagination
		Items []Item `json:"items"`
	}

	// ItemCreationInput represents what a user could set as input for creating items
	ItemCreationInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// ItemUpdateInput represents what a user could set as input for updating items
	ItemUpdateInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// ItemDataManager describes a structure capable of storing items permanently
	ItemDataManager interface {
		GetItem(ctx context.Context, itemID, userID uint64) (*Item, error)
		GetItemCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
		GetAllItemsCount(ctx context.Context) (uint64, error)
		GetItems(ctx context.Context, filter *QueryFilter, userID uint64) (*ItemList, error)
		GetAllItemsForUser(ctx context.Context, userID uint64) ([]Item, error)
		CreateItem(ctx context.Context, input *ItemCreationInput) (*Item, error)
		UpdateItem(ctx context.Context, updated *Item) error
		ArchiveItem(ctx context.Context, id, userID uint64) error
	}

	// ItemDataServer describes a structure capable of serving traffic related to items
	ItemDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		ListHandler() http.HandlerFunc
		CreateHandler() http.HandlerFunc
		ReadHandler() http.HandlerFunc
		UpdateHandler() http.HandlerFunc
		ArchiveHandler() http.HandlerFunc
	}
)

// Update merges an ItemInput with an item
func (x *Item) Update(input *ItemUpdateInput) {
	if input.Name != "" && input.Name != x.Name {
		x.Name = input.Name
	}

	if input.Details != "" && input.Details != x.Details {
		x.Details = input.Details
	}
}

// ToInput creates a ItemUpdateInput struct for an item
func (x *Item) ToInput() *ItemUpdateInput {
	return &ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}
