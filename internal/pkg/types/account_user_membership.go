package types

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
)

type (
	// AccountUserMembership defines a relationship between a user and an account.
	AccountUserMembership struct {
		ID               uint64                      `json:"id"`
		ExternalID       string                      `json:"externalID"`
		BelongsToUser    uint64                      `json:"belongsToUser"`
		BelongsToAccount uint64                      `json:"belongsToAccount"`
		UserPermissions  bitmask.SiteUserPermissions `json:"userPermissions"`
		CreatedOn        uint64                      `json:"createdOn"`
		ArchivedOn       *uint64                     `json:"archivedOn"`
	}

	// AccountUserMembershipList represents a list of account user memberships.
	AccountUserMembershipList struct {
		Pagination
		AccountUserMemberships []*AccountUserMembership `json:"accountUserMemberships"`
	}

	// AccountUserMembershipCreationInput represents what a User could set as input for creating account user memberships.
	AccountUserMembershipCreationInput struct {
		BelongsToUser    uint64                      `json:"belongsToUser"`
		BelongsToAccount uint64                      `json:"belongsToAccount"`
		UserPermissions  bitmask.SiteUserPermissions `json:"userPermissions"`
	}

	// AccountUserMembershipUpdateInput represents what a User could set as input for updating account user memberships.
	AccountUserMembershipUpdateInput struct {
		BelongsToUser    uint64                      `json:"belongsToUser"`
		BelongsToAccount uint64                      `json:"belongsToAccount"`
		UserPermissions  bitmask.SiteUserPermissions `json:"userPermissions"`
	}

	// AccountUserMembershipSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountUserMembershipSQLQueryBuilder interface {
		BuildArchiveAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{})
		BuildGetAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{})
		BuildMarkAccountAsUserDefaultQuery(userID, accountID uint64) (query string, args []interface{})
		BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{})
		BuildCreateMembershipForNewUserQuery(userID, accountID uint64) (query string, args []interface{})
		BuildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{})
		BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{})
	}

	// AccountUserMembershipDataManager describes a structure capable of storing accountUserMemberships permanently.
	AccountUserMembershipDataManager interface {
		MarkAccountAsUserDefault(ctx context.Context, userID, accountID, performedBy uint64) error
		UserIsMemberOfAccount(ctx context.Context, userID, accountID, performedBy uint64) error
		AddUserToAccount(ctx context.Context, userID, accountID, performedBy uint64) error
		RemoveUserFromAccount(ctx context.Context, userID, accountID, performedBy uint64) error
	}
)
