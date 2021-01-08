package types

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
)

type (
	// AccountMembership defines a relationship between a user and an account.
	AccountMembership struct {
		ID                 uint64                         `json:"id"`
		PrimaryUserAccount bool                           `json:"primaryUserAccount"`
		BelongsToAccount   uint64                         `json:"belongsToAccount"`
		BelongsToUser      uint64                         `json:"belongsToUser"`
		UserPermissions    bitmask.AccountUserPermissions `json:"userPermissions"`
		CreatedOn          uint64                         `json:"createdOn"`
		ArchivedOn         *uint64                        `json:"archivedOn"`
	}
)
