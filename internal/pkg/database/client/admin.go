package dbclient

import (
	"context"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	c.querier.LogUserBanEvent(ctx, banGiver, banRecipient)
}
