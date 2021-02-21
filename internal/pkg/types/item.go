package types

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ItemsSearchIndexName is the name of the index used to search through items.
	ItemsSearchIndexName search.IndexName = "items"
)

type (
	// Item represents an item.
	Item struct {
		ID               uint64  `json:"id"`
		ExternalID       string  `json:"externalID"`
		Name             string  `json:"name"`
		Details          string  `json:"details"`
		CreatedOn        uint64  `json:"createdOn"`
		LastUpdatedOn    *uint64 `json:"lastUpdatedOn"`
		ArchivedOn       *uint64 `json:"archivedOn"`
		BelongsToAccount uint64  `json:"belongsToAccount"`
	}

	// ItemList represents a list of items.
	ItemList struct {
		Pagination
		Items []*Item `json:"items"`
	}

	// ItemCreationInput represents what a User could set as input for creating items.
	ItemCreationInput struct {
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount uint64 `json:"-"`
	}

	// ItemUpdateInput represents what a User could set as input for updating items.
	ItemUpdateInput struct {
		Name             string `json:"name"`
		Details          string `json:"details"`
		BelongsToAccount uint64 `json:"-"`
	}

	// ItemSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	ItemSQLQueryBuilder interface {
		BuildItemExistsQuery(itemID, accountID uint64) (query string, args []interface{})
		BuildGetItemQuery(itemID, accountID uint64) (query string, args []interface{})
		BuildGetAllItemsCountQuery() string
		BuildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetItemsQuery(accountID uint64, forAdmin bool, filter *QueryFilter) (query string, args []interface{})
		BuildGetItemsWithIDsQuery(accountID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{})
		BuildCreateItemQuery(input *ItemCreationInput) (query string, args []interface{})
		BuildUpdateItemQuery(input *Item) (query string, args []interface{})
		BuildArchiveItemQuery(itemID, accountID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{})
	}

	// ItemDataManager describes a structure capable of storing items permanently.
	ItemDataManager interface {
		ItemExists(ctx context.Context, itemID, accountID uint64) (bool, error)
		GetItem(ctx context.Context, itemID, accountID uint64) (*Item, error)
		GetAllItemsCount(ctx context.Context) (uint64, error)
		GetAllItems(ctx context.Context, resultChannel chan []*Item, bucketSize uint16) error
		GetItems(ctx context.Context, accountID uint64, filter *QueryFilter) (*ItemList, error)
		GetItemsForAdmin(ctx context.Context, filter *QueryFilter) (*ItemList, error)
		GetItemsWithIDs(ctx context.Context, accountID uint64, limit uint8, ids []uint64) ([]*Item, error)
		GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]*Item, error)
		CreateItem(ctx context.Context, input *ItemCreationInput, createdByUser uint64) (*Item, error)
		UpdateItem(ctx context.Context, updated *Item, changedByUser uint64, changes []FieldChangeSummary) error
		ArchiveItem(ctx context.Context, itemID, belongsToAccount, archivedByUserID uint64) error
		GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]*AuditLogEntry, error)
	}

	// ItemDataService describes a structure capable of serving traffic related to items.
	ItemDataService interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		SearchHandler(res http.ResponseWriter, req *http.Request)
		ListHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ExistenceHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an ItemUpdateInput with an item.
func (x *Item) Update(input *ItemUpdateInput) []FieldChangeSummary {
	var out []FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	if input.Details != "" && input.Details != x.Details {
		out = append(out, FieldChangeSummary{
			FieldName: "Details",
			OldValue:  x.Details,
			NewValue:  input.Details,
		})

		x.Details = input.Details
	}

	return out
}

// Validate validates a ItemCreationInput.
func (x *ItemCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

// Validate validates a ItemUpdateInput.
func (x *ItemUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
