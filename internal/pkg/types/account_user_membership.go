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
		BuildGetAccountUserMembershipQuery(accountUserMembershipID, userID uint64) (query string, args []interface{})
		BuildGetAllAccountUserMembershipsCountQuery() string
		BuildGetBatchOfAccountUserMembershipsQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetAccountUserMembershipsQuery(userID uint64, forAdmin bool, filter *QueryFilter) (query string, args []interface{})
		BuildCreateAccountUserMembershipQuery(input *AccountUserMembershipCreationInput) (query string, args []interface{})
		BuildArchiveAccountUserMembershipQuery(accountUserMembershipID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAccountUserMembershipQuery(accountUserMembershipID uint64) (query string, args []interface{})
	}

	// AccountUserMembershipDataManager describes a structure capable of storing accountUserMemberships permanently.
	AccountUserMembershipDataManager interface {
		GetAccountUserMembership(ctx context.Context, accountUserMembershipID, accountID uint64) (*AccountUserMembership, error)
		GetAllAccountUserMembershipsCount(ctx context.Context) (uint64, error)
		GetAllAccountUserMemberships(ctx context.Context, resultChannel chan []*AccountUserMembership, bucketSize uint16) error
		GetAccountUserMemberships(ctx context.Context, accountID uint64, filter *QueryFilter) (*AccountUserMembershipList, error)
		GetAccountUserMembershipsForAdmin(ctx context.Context, filter *QueryFilter) (*AccountUserMembershipList, error)
		CreateAccountUserMembership(ctx context.Context, input *AccountUserMembershipCreationInput) (*AccountUserMembership, error)
		ArchiveAccountUserMembership(ctx context.Context, accountUserMembershipID, accountID uint64) error
	}

	// AccountUserMembershipAuditManager describes a structure capable of .
	AccountUserMembershipAuditManager interface {
		GetAuditLogEntriesForAccountUserMembership(ctx context.Context, itemID uint64) ([]*AuditLogEntry, error)
		LogAccountUserMembershipCreationEvent(ctx context.Context, item *AccountUserMembership)
		LogAccountUserMembershipUpdateEvent(ctx context.Context, userID, itemID uint64, changes []FieldChangeSummary)
		LogAccountUserMembershipArchiveEvent(ctx context.Context, userID, itemID uint64)
	}
)
