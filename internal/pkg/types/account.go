package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Account represents an account.
	Account struct {
		ID              uint64  `json:"id"`
		Name            string  `json:"name"`
		CreatedOn       uint64  `json:"createdOn"`
		LastUpdatedOn   *uint64 `json:"lastUpdatedOn"`
		ArchivedOn      *uint64 `json:"archivedOn"`
		PlanID          *uint64 `json:"planID"`
		PersonalAccount bool    `json:"personalAccount"`
		BelongsToUser   uint64  `json:"belongsToUser"`
	}

	// AccountList represents a list of accounts.
	AccountList struct {
		Pagination
		Accounts []*Account `json:"accounts"`
	}

	// AccountCreationInput represents what a User could set as input for creating accounts.
	AccountCreationInput struct {
		Name          string `json:"name"`
		BelongsToUser uint64 `json:"-"`
	}

	// AccountUpdateInput represents what a User could set as input for updating accounts.
	AccountUpdateInput struct {
		Name          string `json:"name"`
		BelongsToUser uint64 `json:"-"`
	}

	// AccountSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountSQLQueryBuilder interface {
		BuildGetAccountQuery(accountID, userID uint64) (query string, args []interface{})
		BuildGetAllAccountsCountQuery() string
		BuildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetAccountsQuery(userID uint64, forAdmin bool, filter *QueryFilter) (query string, args []interface{})
		BuildCreateAccountQuery(input *AccountCreationInput) (query string, args []interface{})
		BuildUpdateAccountQuery(input *Account) (query string, args []interface{})
		BuildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{})
	}

	// AccountDataManager describes a structure capable of storing accounts permanently.
	AccountDataManager interface {
		GetAccount(ctx context.Context, accountID, userID uint64) (*Account, error)
		GetAllAccountsCount(ctx context.Context) (uint64, error)
		GetAllAccounts(ctx context.Context, resultChannel chan []*Account, bucketSize uint16) error
		GetAccounts(ctx context.Context, userID uint64, filter *QueryFilter) (*AccountList, error)
		GetAccountsForAdmin(ctx context.Context, filter *QueryFilter) (*AccountList, error)
		CreateAccount(ctx context.Context, input *AccountCreationInput) (*Account, error)
		UpdateAccount(ctx context.Context, updated *Account) error
		ArchiveAccount(ctx context.Context, accountID, userID uint64) error
	}

	// AccountAuditManager describes a structure capable of .
	AccountAuditManager interface {
		GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*AuditLogEntry, error)
		LogAccountCreationEvent(ctx context.Context, account *Account)
		LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []FieldChangeSummary)
		LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64)
	}

	// AccountDataService describes a structure capable of serving traffic related to accounts.
	AccountDataService interface {
		UpdateInputMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an AccountUpdateInput with an account.
func (x *Account) Update(input *AccountUpdateInput) []FieldChangeSummary {
	var out []FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	return out
}

// Validate validates a AccountCreationInput.
func (x *AccountCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

// Validate validates a AccountUpdateInput.
func (x *AccountUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
