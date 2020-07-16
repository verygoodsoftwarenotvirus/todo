package models

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"
)

const (
	// ItemsSearchIndexName is the name of the index used to search through items.
	ItemsSearchIndexName search.IndexName = "items"
)

type (
	// Item represents an item.
	Item struct {
		ID            uint64  `json:"id"`
		Name          string  `json:"name"`
		Details       string  `json:"details"`
		CreatedOn     uint64  `json:"createdOn"`
		UpdatedOn     *uint64 `json:"updatedOn"`
		ArchivedOn    *uint64 `json:"archivedOn"`
		BelongsToUser uint64  `json:"belongsToUser"`
	}

	// ItemList represents a list of items.
	ItemList struct {
		Pagination
		Items []Item `json:"items"`
	}

	// ItemCreationInput represents what a user could set as input for creating items.
	ItemCreationInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// ItemUpdateInput represents what a user could set as input for updating items.
	ItemUpdateInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// ItemDataManager describes a structure capable of storing items permanently.
	ItemDataManager interface {
		ItemExists(ctx context.Context, itemID, userID uint64) (bool, error)
		GetItem(ctx context.Context, itemID, userID uint64) (*Item, error)
		GetAllItemsCount(ctx context.Context) (uint64, error)
		GetAllItems(ctx context.Context, resultChannel chan []Item) error
		GetItems(ctx context.Context, userID uint64, filter *QueryFilter) (*ItemList, error)
		GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]Item, error)
		CreateItem(ctx context.Context, input *ItemCreationInput) (*Item, error)
		UpdateItem(ctx context.Context, updated *Item) error
		ArchiveItem(ctx context.Context, itemID, userID uint64) error
	}

	// ItemDataServer describes a structure capable of serving traffic related to items.
	ItemDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		SearchHandler() http.HandlerFunc
		ListHandler() http.HandlerFunc
		CreateHandler() http.HandlerFunc
		ExistenceHandler() http.HandlerFunc
		ReadHandler() http.HandlerFunc
		UpdateHandler() http.HandlerFunc
		ArchiveHandler() http.HandlerFunc
	}
)

// Update merges an ItemInput with an item.
func (x *Item) Update(input *ItemUpdateInput) {
	if input.Name != "" && input.Name != x.Name {
		x.Name = input.Name
	}

	if input.Details != "" && input.Details != x.Details {
		x.Details = input.Details
	}
}

// ToUpdateInput creates a ItemUpdateInput struct for an item.
func (x *Item) ToUpdateInput() *ItemUpdateInput {
	return &ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}
