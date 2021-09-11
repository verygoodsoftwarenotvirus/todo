package types

import (
	"context"
	"encoding/gob"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ItemsSearchIndexName is the name of the index used to search through items.
	ItemsSearchIndexName search.IndexName = "items"
)

func init() {
	gob.Register(new(Item))
	gob.Register(new(ItemList))
	gob.Register(new(ItemCreationInput))
	gob.Register(new(ItemUpdateInput))
}

type (
	// Item represents an item.
	Item struct {
		ArchivedOn       *uint64 `json:"archivedOn"`
		LastUpdatedOn    *uint64 `json:"lastUpdatedOn"`
		Name             string  `json:"name"`
		Details          string  `json:"details"`
		ID               string  `json:"id"`
		BelongsToAccount string  `json:"belongsToAccount"`
		CreatedOn        uint64  `json:"createdOn"`
	}

	// ItemList represents a list of items.
	ItemList struct {
		Items []*Item `json:"items"`
		Pagination
	}

	// ItemCreationInput represents what a user could set as input for creating items.
	ItemCreationInput struct {
		ID               string `json:"-"`
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"-"`
	}

	// ItemDatabaseCreationInput represents what a user could set as input for creating items.
	ItemDatabaseCreationInput struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"belongsToAccount"`
	}

	// ItemUpdateInput represents what a user could set as input for updating items.
	ItemUpdateInput struct {
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"-"`
	}

	// ItemDataManager describes a structure capable of storing items permanently.
	ItemDataManager interface {
		ItemExists(ctx context.Context, itemID, accountID string) (bool, error)
		GetItem(ctx context.Context, itemID, accountID string) (*Item, error)
		GetAllItemsCount(ctx context.Context) (uint64, error)
		GetAllItems(ctx context.Context, resultChannel chan []*Item, bucketSize uint16) error
		GetItems(ctx context.Context, accountID string, filter *QueryFilter) (*ItemList, error)
		GetItemsWithIDs(ctx context.Context, accountID string, limit uint8, ids []string) ([]*Item, error)
		CreateItem(ctx context.Context, input *ItemDatabaseCreationInput, createdByUser string) (*Item, error)
		UpdateItem(ctx context.Context, updated *Item, changedByUser string, changes []*FieldChangeSummary) error
		ArchiveItem(ctx context.Context, itemID, accountID, archivedBy string) error
		GetAuditLogEntriesForItem(ctx context.Context, itemID string) ([]*AuditLogEntry, error)
	}

	// ItemDataService describes a structure capable of serving traffic related to items.
	ItemDataService interface {
		SearchHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ExistenceHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an ItemUpdateInput with an item.
func (x *Item) Update(input *ItemUpdateInput) []*FieldChangeSummary {
	var out []*FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, &FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	if input.Details != "" && input.Details != x.Details {
		out = append(out, &FieldChangeSummary{
			FieldName: "Details",
			OldValue:  x.Details,
			NewValue:  input.Details,
		})

		x.Details = input.Details
	}

	return out
}

var _ validation.ValidatableWithContext = (*ItemCreationInput)(nil)

// ValidateWithContext validates a ItemCreationInput.
func (x *ItemCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ItemDatabaseCreationInput)(nil)

// ValidateWithContext validates a ItemDatabaseCreationInput.
func (x *ItemDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.BelongsToAccount, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ItemUpdateInput)(nil)

// ValidateWithContext validates a ItemUpdateInput.
func (x *ItemUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
