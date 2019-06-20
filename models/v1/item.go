package models

import (
	"context"
	"net/http"
)

type (
	// Item represents an item
	Item struct {
		ID         uint64  `json:"id"`
		Name       string  `json:"name"`
		Details    string  `json:"details"`
		CreatedOn  uint64  `json:"created_on"`
		UpdatedOn  *uint64 `json:"updated_on"`
		ArchivedOn *uint64 `json:"archived_on"`
		BelongsTo  uint64  `json:"belongs_to"`
	}

	// ItemList represents a list of items
	ItemList struct {
		Pagination
		Items []Item `json:"items"`
	}

	// ItemCreationInput represents what a user could set as input for items
	ItemCreationInput struct {
		Name      string `json:"name"`
		Details   string `json:"details"`
		BelongsTo uint64 `json:"-"`
	}

	// ItemUpdateInput represents what a user could set as input for items
	ItemUpdateInput struct {
		Name      string `json:"name"`
		Details   string `json:"details"`
		BelongsTo uint64 `json:"-"`
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

		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an ItemInput with an Item
func (i *Item) Update(input *ItemUpdateInput) {
	if input.Name != "" || input.Name != i.Name {
		i.Name = input.Name
	}

	if input.Details != "" || input.Details != i.Details {
		i.Details = input.Details
	}
}
