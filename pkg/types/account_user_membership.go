package types

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// AccountUserMembership defines a relationship between a user and an account.
	AccountUserMembership struct {
		ArchivedOn             *uint64                           `json:"archivedOn"`
		LastUpdatedOn          *uint64                           `json:"lastUpdatedOn"`
		BelongsToUser          uint64                            `json:"belongsToUser"`
		BelongsToAccount       uint64                            `json:"belongsToAccount"`
		CreatedOn              uint64                            `json:"createdOn"`
		ID                     uint64                            `json:"id"`
		UserAccountPermissions permissions.ServiceUserPermission `json:"userAccountPermissions"`
		DefaultAccount         bool                              `json:"defaultAccount"`
	}

	// AccountUserMembershipList represents a list of account user memberships.
	AccountUserMembershipList struct {
		AccountUserMemberships []*AccountUserMembership `json:"accountUserMemberships"`
		Pagination
	}

	// AccountUserMembershipCreationInput represents what a User could set as input for creating account user memberships.
	AccountUserMembershipCreationInput struct {
		BelongsToUser          uint64                            `json:"belongsToUser"`
		BelongsToAccount       uint64                            `json:"belongsToAccount"`
		UserAccountPermissions permissions.ServiceUserPermission `json:"userAccountPermissions"`
	}

	// AccountUserMembershipUpdateInput represents what a User could set as input for updating account user memberships.
	AccountUserMembershipUpdateInput struct {
		BelongsToUser          uint64                            `json:"belongsToUser"`
		BelongsToAccount       uint64                            `json:"belongsToAccount"`
		UserAccountPermissions permissions.ServiceUserPermission `json:"userAccountPermissions"`
	}

	// AddUserToAccountInput represents what a User could set as input for updating account user memberships.
	AddUserToAccountInput struct {
		Reason                 string                            `json:"reason"`
		UserID                 uint64                            `json:"userID"`
		AccountID              uint64                            `json:"accountID"`
		UserAccountPermissions permissions.ServiceUserPermission `json:"userAccountPermissions"`
	}

	// AccountOwnershipTransferInput represents what a User could set as input for updating account user memberships.
	AccountOwnershipTransferInput struct {
		Reason       string `json:"reason"`
		CurrentOwner uint64 `json:"currentOwner"`
		NewOwner     uint64 `json:"newOwner"`
	}

	// ModifyUserPermissionsInput  represents what a User could set as input for updating account user memberships.
	ModifyUserPermissionsInput struct {
		Reason                 string                            `json:"reason"`
		UserAccountPermissions permissions.ServiceUserPermission `json:"userAccountPermissions"`
	}

	// AccountUserMembershipDataManager describes a structure capable of storing accountUserMemberships permanently.
	AccountUserMembershipDataManager interface {
		BuildSessionContextDataForUser(ctx context.Context, userID uint64) (*SessionContextData, error)
		GetDefaultAccountIDForUser(ctx context.Context, userID uint64) (uint64, error)
		MarkAccountAsUserDefault(ctx context.Context, userID, accountID, changedByUser uint64) error
		UserIsMemberOfAccount(ctx context.Context, userID, accountID uint64) (bool, error)
		ModifyUserPermissions(ctx context.Context, accountID, userID, changedByUser uint64, input *ModifyUserPermissionsInput) error
		TransferAccountOwnership(ctx context.Context, accountID uint64, transferredBy uint64, input *AccountOwnershipTransferInput) error
		AddUserToAccount(ctx context.Context, input *AddUserToAccountInput, addedByUser uint64) error
		RemoveUserFromAccount(ctx context.Context, userID, accountID, removedByUser uint64, reason string) error
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
		validation.Field(&x.UserAccountPermissions, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}
