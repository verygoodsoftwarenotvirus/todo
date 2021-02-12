package querier

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*Client)(nil)
)

// MarkAccountAsUserDefault does a thing.
func (c *Client) MarkAccountAsUserDefault(ctx context.Context, userID, accountID, performedBy uint64) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.PerformedByKey: performedBy,
	})

	logger.Debug("MarkAccountAsUserDefault called")

	return nil
}

// UserIsMemberOfAccount does a thing.
func (c *Client) UserIsMemberOfAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.PerformedByKey: performedBy,
	})

	logger.Debug("UserIsMemberOfAccount called")

	return nil
}

// AddUserToAccount does a thing.
func (c *Client) AddUserToAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.PerformedByKey: performedBy,
	})

	logger.Debug("AddUserToAccount called")

	return nil
}

// RemoveUserFromAccount does a thing.
func (c *Client) RemoveUserFromAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.PerformedByKey: performedBy,
	})

	logger.Debug("RemoveUserFromAccount called")

	return nil
}

// GetAuditLogEntriesForAccountUserMembership fetches a list of audit log entries from the database that relate to a given accountUserMembership.
func (c *Client) GetAuditLogEntriesForAccountUserMembership(ctx context.Context, accountUserMembershipID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForAccountUserMembership called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountUserMembershipQuery(accountUserMembershipID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning audit log entries: %w", err)
	}

	return auditLogEntries, nil
}
