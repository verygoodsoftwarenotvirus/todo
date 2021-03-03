package types

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

type (
	// AccountUserMembership defines a relationship between a user and an account.
	AccountUserMembership struct {
		ID                     uint64                             `json:"id"`
		BelongsToUser          uint64                             `json:"belongsToUser"`
		BelongsToAccount       uint64                             `json:"belongsToAccount"`
		DefaultAccount         bool                               `json:"defaultAccount"`
		UserAccountPermissions permissions.ServiceUserPermissions `json:"userAccountPermissions"`
		CreatedOn              uint64                             `json:"createdOn"`
		ArchivedOn             *uint64                            `json:"archivedOn"`
	}

	// AccountUserMembershipList represents a list of account user memberships.
	AccountUserMembershipList struct {
		Pagination
		AccountUserMemberships []*AccountUserMembership `json:"accountUserMemberships"`
	}

	// AccountUserMembershipCreationInput represents what a User could set as input for creating account user memberships.
	AccountUserMembershipCreationInput struct {
		BelongsToUser          uint64                             `json:"belongsToUser"`
		BelongsToAccount       uint64                             `json:"belongsToAccount"`
		UserAccountPermissions permissions.ServiceUserPermissions `json:"userAccountPermissions"`
	}

	// AccountUserMembershipUpdateInput represents what a User could set as input for updating account user memberships.
	AccountUserMembershipUpdateInput struct {
		BelongsToUser          uint64                             `json:"belongsToUser"`
		BelongsToAccount       uint64                             `json:"belongsToAccount"`
		UserAccountPermissions permissions.ServiceUserPermissions `json:"userAccountPermissions"`
	}

	// AddUserToAccountInput represents what a User could set as input for updating account user memberships.
	AddUserToAccountInput struct {
		UserID                 uint64                             `json:"userID"`
		AccountID              uint64                             `json:"-"`
		UserAccountPermissions permissions.ServiceUserPermissions `json:"userAccountPermissions"`
		Reason                 string                             `json:"reason"`
	}

	// TransferAccountOwnershipInput represents what a User could set as input for updating account user memberships.
	TransferAccountOwnershipInput struct {
		CurrentOwner uint64 `json:"currentOwner"`
		NewOwner     uint64 `json:"newOwner"`
		Reason       string `json:"reason"`
	}

	// ModifyUserPermissionsInput  represents what a User could set as input for updating account user memberships.
	ModifyUserPermissionsInput struct {
		UserID                 uint64                             `json:"userID"`
		UserAccountPermissions permissions.ServiceUserPermissions `json:"userAccountPermissions"`
		Reason                 string                             `json:"reason"`
	}

	// RemoveUserFromAccountInput represents what a User could set as input for updating account user memberships.
	RemoveUserFromAccountInput struct {
		UserID uint64 `json:"userID"`
		Reason string `json:"reason"`
	}

	// AccountUserMembershipSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountUserMembershipSQLQueryBuilder interface {
		BuildArchiveAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{})
		BuildGetAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{})
		BuildMarkAccountAsUserDefaultQuery(userID, accountID uint64) (query string, args []interface{})
		BuildModifyUserPermissionsQuery(userID, accountID uint64, permissions permissions.ServiceUserPermissions) (query string, args []interface{})
		BuildTransferAccountOwnershipQuery(oldOwnerID, newOwnerID, accountID uint64) (query string, args []interface{})
		BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{})
		BuildCreateMembershipForNewUserQuery(userID, accountID uint64) (query string, args []interface{})
		BuildAddUserToAccountQuery(input *AddUserToAccountInput) (query string, args []interface{})
		BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{})
	}

	// AccountUserMembershipDataManager describes a structure capable of storing accountUserMemberships permanently.
	AccountUserMembershipDataManager interface {
		GetMembershipsForUser(ctx context.Context, userID uint64) (uint64, map[uint64]permissions.ServiceUserPermissions, error)
		MarkAccountAsUserDefault(ctx context.Context, userID, accountID, changedByUser uint64) error
		UserIsMemberOfAccount(ctx context.Context, userID, accountID uint64) (bool, error)
		ModifyUserPermissions(ctx context.Context, accountID, changedByUser uint64, input *ModifyUserPermissionsInput) error
		TransferAccountOwnership(ctx context.Context, accountID uint64, transferredBy uint64, input *TransferAccountOwnershipInput) error
		AddUserToAccount(ctx context.Context, input *AddUserToAccountInput, addedByUser uint64) error
		RemoveUserFromAccount(ctx context.Context, userID, accountID, removedByUser uint64, reason string) error
	}
)

// Validate validates an AddUserToAccountInput.
func (x *AddUserToAccountInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.UserID, validation.Required),
	)
}

// Validate validates a TransferAccountOwnershipInput.
func (x *TransferAccountOwnershipInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.CurrentOwner, validation.Required),
		validation.Field(&x.NewOwner, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}

// Validate validates a ModifyUserPermissionsInput.
func (x *ModifyUserPermissionsInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.UserID, validation.Required),
		validation.Field(&x.UserAccountPermissions, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}

// Validate validates a RemoveUserFromAccountInput.
func (x *RemoveUserFromAccountInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.UserID, validation.Required),
		validation.Field(&x.Reason, validation.Required),
	)
}
