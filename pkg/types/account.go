package types

import (
	"context"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PaidAccountBillingStatus indicates an account is fully paid.
	PaidAccountBillingStatus AccountBillingStatus = "paid"
	// UnpaidAccountBillingStatus indicates an account is not paid.
	UnpaidAccountBillingStatus AccountBillingStatus = "unpaid"
)

type (
	// AccountBillingStatus is the type to use/compare against when checking billing status.
	AccountBillingStatus string

	// Account represents an account.
	Account struct {
		ArchivedOn                 *uint64                  `json:"archivedOn"`
		SubscriptionPlanID         *uint64                  `json:"subscriptionPlanID"`
		LastUpdatedOn              *uint64                  `json:"lastUpdatedOn"`
		Name                       string                   `json:"name"`
		BillingStatus              AccountBillingStatus     `json:"billingStatus"`
		ContactEmail               string                   `json:"contactEmail"`
		ContactPhone               string                   `json:"contactPhone"`
		PaymentProcessorCustomerID string                   `json:"paymentProcessorCustomer"`
		ExternalID                 string                   `json:"externalID"`
		Members                    []*AccountUserMembership `json:"members"`
		CreatedOn                  uint64                   `json:"createdOn"`
		ID                         uint64                   `json:"id"`
		BelongsToUser              uint64                   `json:"belongsToUser"`
	}

	// AccountList represents a list of accounts.
	AccountList struct {
		Accounts []*Account `json:"accounts"`
		Pagination
	}

	// AccountCreationInput represents what a User could set as input for creating accounts.
	AccountCreationInput struct {
		Name          string `json:"name"`
		ContactEmail  string `json:"contactEmail"`
		ContactPhone  string `json:"contactPhone"`
		BelongsToUser uint64 `json:"-"`
	}

	// AccountUpdateInput represents what a User could set as input for updating accounts.
	AccountUpdateInput struct {
		Name          string `json:"name"`
		ContactEmail  string `json:"contactEmail"`
		ContactPhone  string `json:"contactPhone"`
		BelongsToUser uint64 `json:"-"`
	}

	// AccountDataManager describes a structure capable of storing accounts permanently.
	AccountDataManager interface {
		GetAccount(ctx context.Context, accountID, userID uint64) (*Account, error)
		GetAllAccountsCount(ctx context.Context) (uint64, error)
		GetAllAccounts(ctx context.Context, resultChannel chan []*Account, bucketSize uint16) error
		GetAccounts(ctx context.Context, userID uint64, filter *QueryFilter) (*AccountList, error)
		GetAccountsForAdmin(ctx context.Context, filter *QueryFilter) (*AccountList, error)
		CreateAccount(ctx context.Context, input *AccountCreationInput, createdByUser uint64) (*Account, error)
		UpdateAccount(ctx context.Context, updated *Account, changedByUser uint64, changes []*FieldChangeSummary) error
		ArchiveAccount(ctx context.Context, accountID, userID, archivedByUser uint64) error
		GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*AuditLogEntry, error)
	}

	// AccountDataService describes a structure capable of serving traffic related to accounts.
	AccountDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
		AddMemberHandler(res http.ResponseWriter, req *http.Request)
		RemoveMemberHandler(res http.ResponseWriter, req *http.Request)
		MarkAsDefaultAccountHandler(res http.ResponseWriter, req *http.Request)
		ModifyMemberPermissionsHandler(res http.ResponseWriter, req *http.Request)
		TransferAccountOwnershipHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an AccountUpdateInput with an account.
func (x *Account) Update(input *AccountUpdateInput) []*FieldChangeSummary {
	var out []*FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, &FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	return out
}

var _ validation.ValidatableWithContext = (*AccountCreationInput)(nil)

// ValidateWithContext validates a AccountCreationInput.
func (x *AccountCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AccountUpdateInput)(nil)

// ValidateWithContext validates a AccountUpdateInput.
func (x *AccountUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

// AccountCreationInputForNewUser creates a new AccountInputCreation struct for a given user.
func AccountCreationInputForNewUser(u *User) *AccountCreationInput {
	return &AccountCreationInput{
		Name:          fmt.Sprintf("%s_default", u.Username),
		BelongsToUser: u.ID,
	}
}
