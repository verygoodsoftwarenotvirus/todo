package types

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// UserMembershipDataType indicates an event is user membership-related.
	UserMembershipDataType dataType = "user_membership"
)

type (
	// AccountUserMembership defines a relationship between a user and an account.
	AccountUserMembership struct {
		_ struct{}

		ArchivedOn       *uint64  `json:"archivedOn"`
		LastUpdatedOn    *uint64  `json:"lastUpdatedOn"`
		ID               string   `json:"id"`
		BelongsToUser    string   `json:"belongsToUser"`
		BelongsToAccount string   `json:"belongsToAccount"`
		AccountRoles     []string `json:"accountRole"`
		CreatedOn        uint64   `json:"createdOn"`
		DefaultAccount   bool     `json:"defaultAccount"`
	}

	// AccountUserMembershipList represents a list of account user memberships.
	AccountUserMembershipList struct {
		_ struct{}

		AccountUserMemberships []*AccountUserMembership `json:"accountUserMemberships"`
		Pagination
	}

	// AccountUserMembershipCreationInput represents what a User could set as input for creating account user memberships.
	AccountUserMembershipCreationInput struct {
		_ struct{}

		ID               string `json:"-"`
		BelongsToUser    string `json:"belongsToUser"`
		BelongsToAccount string `json:"belongsToAccount"`
	}

	// AccountUserMembershipUpdateInput represents what a User could set as input for updating account user memberships.
	AccountUserMembershipUpdateInput struct {
		_ struct{}

		BelongsToUser    string `json:"belongsToUser"`
		BelongsToAccount string `json:"belongsToAccount"`
	}

	// AddUserToAccountInput represents what a User could set as input for updating account user memberships.
	AddUserToAccountInput struct {
		_ struct{}

		ID           string   `json:"-"`
		Reason       string   `json:"reason"`
		UserID       string   `json:"userID"`
		AccountID    string   `json:"accountID"`
		AccountRoles []string `json:"accountRole"`
	}

	// AccountOwnershipTransferInput represents what a User could set as input for updating account user memberships.
	AccountOwnershipTransferInput struct {
		_ struct{}

		Reason       string `json:"reason"`
		CurrentOwner string `json:"currentOwner"`
		NewOwner     string `json:"newOwner"`
	}

	// ModifyUserPermissionsInput  represents what a User could set as input for updating account user memberships.
	ModifyUserPermissionsInput struct {
		_ struct{}

		Reason   string   `json:"reason"`
		NewRoles []string `json:"newRoles"`
	}

	// AccountUserMembershipDataManager describes a structure capable of storing accountUserMemberships permanently.
	AccountUserMembershipDataManager interface {
		BuildSessionContextDataForUser(ctx context.Context, userID string) (*SessionContextData, error)
		GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error)
		MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error
		UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error)
		ModifyUserPermissions(ctx context.Context, accountID, userID string, input *ModifyUserPermissionsInput) error
		TransferAccountOwnership(ctx context.Context, accountID string, input *AccountOwnershipTransferInput) error
		AddUserToAccount(ctx context.Context, input *AddUserToAccountInput) error
		RemoveUserFromAccount(ctx context.Context, userID, accountID string) error
	}
)

var _ validation.ValidatableWithContext = (*AddUserToAccountInput)(nil)

// ValidateWithContext validates an AddUserToAccountInput.
func (x *AddUserToAccountInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.UserID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AccountOwnershipTransferInput)(nil)

// ValidateWithContext validates a AccountOwnershipTransferInput.
func (x *AccountOwnershipTransferInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.CurrentOwner, validation.Required),
		validation.Field(&x.NewOwner, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ModifyUserPermissionsInput)(nil)

// ValidateWithContext validates a ModifyUserPermissionsInput.
func (x *ModifyUserPermissionsInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.NewRoles, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}
