package types

import (
	"context"
	"encoding/gob"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ItemDataType indicates an event is related to an item.
	ItemDataType dataType = "item"
)

func init() {
	gob.Register(new(Item))
	gob.Register(new(ItemList))
	gob.Register(new(ItemCreationRequestInput))
	gob.Register(new(ItemUpdateRequestInput))
}

type (
	// Item represents an item.
	Item struct {
		_                struct{}
		LastUpdatedOn    *uint64 `json:"lastUpdatedOn"`
		ArchivedOn       *uint64 `json:"archivedOn"`
		ID               string  `json:"id"`
		Name             string  `json:"name"`
		Details          string  `json:"details"`
		BelongsToAccount string  `json:"belongsToAccount"`
		CreatedOn        uint64  `json:"createdOn"`
	}

	// ItemList represents a list of items.
	ItemList struct {
		_     struct{}
		Items []*Item `json:"items"`
		Pagination
	}

	// ItemCreationRequestInput represents what a user could set as input for creating items.
	ItemCreationRequestInput struct {
		_ struct{}

		ID               string `json:"-"`
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"-"`
	}

	// ItemDatabaseCreationInput represents what a user could set as input for creating items.
	ItemDatabaseCreationInput struct {
		_ struct{}

		ID               string `json:"id"`
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"belongsToAccount"`
	}

	// ItemUpdateRequestInput represents what a user could set as input for updating items.
	ItemUpdateRequestInput struct {
		_ struct{}

		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount string `json:"-"`
	}

	// ItemDataManager describes a structure capable of storing items permanently.
	ItemDataManager interface {
		ItemExists(ctx context.Context, itemID, accountID string) (bool, error)
		GetItem(ctx context.Context, itemID, accountID string) (*Item, error)
		GetTotalItemCount(ctx context.Context) (uint64, error)
		GetItems(ctx context.Context, accountID string, filter *QueryFilter) (*ItemList, error)
		GetItemsWithIDs(ctx context.Context, accountID string, limit uint8, ids []string) ([]*Item, error)
		CreateItem(ctx context.Context, input *ItemDatabaseCreationInput) (*Item, error)
		UpdateItem(ctx context.Context, updated *Item) error
		ArchiveItem(ctx context.Context, itemID, accountID string) error
	}

	// ItemDataService describes a structure capable of serving traffic related to items.
	ItemDataService interface {
		SearchHandler(res http.ResponseWriter, req *http.Request)
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an ItemUpdateRequestInput with an item.
func (x *Item) Update(input *ItemUpdateRequestInput) {
	if input.Name != "" && input.Name != x.Name {
		x.Name = input.Name
	}

	if input.Details != "" && input.Details != x.Details {
		x.Details = input.Details
	}
}

var _ validation.ValidatableWithContext = (*ItemCreationRequestInput)(nil)

// ValidateWithContext validates a ItemCreationRequestInput.
func (x *ItemCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Details, validation.Required),
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
		validation.Field(&x.Details, validation.Required),
		validation.Field(&x.BelongsToAccount, validation.Required),
	)
}

// ItemDatabaseCreationInputFromItemCreationInput creates a DatabaseCreationInput from a CreationInput.
func ItemDatabaseCreationInputFromItemCreationInput(input *ItemCreationRequestInput) *ItemDatabaseCreationInput {
	x := &ItemDatabaseCreationInput{}

	x.Name = input.Name
	x.Details = input.Details

	return x
}

var _ validation.ValidatableWithContext = (*ItemUpdateRequestInput)(nil)

// ValidateWithContext validates a ItemUpdateRequestInput.
func (x *ItemUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Details, validation.Required),
	)
}
