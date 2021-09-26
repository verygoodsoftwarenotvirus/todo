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
		_ struct{}

		ArchivedOn                 *uint64                  `json:"archivedOn"`
		SubscriptionPlanID         *uint64                  `json:"subscriptionPlanID"`
		LastUpdatedOn              *uint64                  `json:"lastUpdatedOn"`
		Name                       string                   `json:"name"`
		BillingStatus              AccountBillingStatus     `json:"billingStatus"`
		ContactEmail               string                   `json:"contactEmail"`
		ContactPhone               string                   `json:"contactPhone"`
		PaymentProcessorCustomerID string                   `json:"paymentProcessorCustomer"`
		BelongsToUser              string                   `json:"belongsToUser"`
		ID                         string                   `json:"id"`
		Members                    []*AccountUserMembership `json:"members"`
		CreatedOn                  uint64                   `json:"createdOn"`
	}

	// AccountList represents a list of accounts.
	AccountList struct {
		_ struct{}

		Accounts []*Account `json:"accounts"`
		Pagination
	}

	// AccountCreationInput represents what a User could set as input for creating accounts.
	AccountCreationInput struct {
		_ struct{}

		ID            string `json:"-"`
		Name          string `json:"name"`
		ContactEmail  string `json:"contactEmail"`
		ContactPhone  string `json:"contactPhone"`
		BelongsToUser string `json:"-"`
	}

	// AccountUpdateInput represents what a User could set as input for updating accounts.
	AccountUpdateInput struct {
		_ struct{}

		Name          string `json:"name"`
		ContactEmail  string `json:"contactEmail"`
		ContactPhone  string `json:"contactPhone"`
		BelongsToUser string `json:"-"`
	}

	// AccountDataManager describes a structure capable of storing accounts permanently.
	AccountDataManager interface {
		GetAccount(ctx context.Context, accountID, userID string) (*Account, error)
		GetAllAccountsCount(ctx context.Context) (uint64, error)
		GetAccounts(ctx context.Context, userID string, filter *QueryFilter) (*AccountList, error)
		GetAccountsForAdmin(ctx context.Context, filter *QueryFilter) (*AccountList, error)
		CreateAccount(ctx context.Context, input *AccountCreationInput) (*Account, error)
		UpdateAccount(ctx context.Context, updated *Account) error
		ArchiveAccount(ctx context.Context, accountID string, userID string) error
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
	}
)

// Update merges an AccountUpdateInput with an account.
func (x *Account) Update(input *AccountUpdateInput) {
	if input.Name != "" && input.Name != x.Name {
		x.Name = input.Name
	}
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
