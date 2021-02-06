package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager  = (*Client)(nil)
	_ types.AccountUserMembershipAuditManager = (*Client)(nil)
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (c *Client) scanAccountUserMembership(scan database.Scanner, includeCounts bool) (x *types.AccountUserMembership, filteredCount, totalCount uint64, err error) {
	x = &types.AccountUserMembership{}

	targetVars := []interface{}{
		&x.ID,
		&x.ExternalID,
		&x.BelongsToUser,
		&x.BelongsToAccount,
		&x.UserPermissions,
		&x.CreatedOn,
		&x.ArchivedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	return x, filteredCount, totalCount, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of account user memberships.
func (c *Client) scanAccountUserMemberships(rows database.ResultIterator, includeCounts bool) (accountUserMemberships []*types.AccountUserMembership, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := c.scanAccountUserMembership(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		accountUserMemberships = append(accountUserMemberships, x)
	}

	if handleErr := c.handleRows(rows); handleErr != nil {
		return nil, 0, 0, handleErr
	}

	return accountUserMemberships, filteredCount, totalCount, nil
}

// GetAccountUserMembership fetches an account user membership from the database.
func (c *Client) GetAccountUserMembership(ctx context.Context, accountUserMembershipID, userID uint64) (*types.AccountUserMembership, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountUserMembershipIDToSpan(span, accountUserMembershipID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.AccountUserMembershipIDKey: accountUserMembershipID,
		keys.UserIDKey:                  userID,
	}).Debug("GetAccountUserMembership called")

	query, args := c.sqlQueryBuilder.BuildGetAccountUserMembershipQuery(accountUserMembershipID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	accountUserMembership, _, _, err := c.scanAccountUserMembership(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning accountUserMembership: %w", err)
	}

	return accountUserMembership, nil
}

// GetAllItemsCount fetches the count of account user memberships from the database that meet a particular filter.
func (c *Client) GetAllAccountUserMembershipsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountUserMembershipsCount called")

	if err = c.db.QueryRowContext(ctx, c.sqlQueryBuilder.BuildGetAllAccountUserMembershipsCountQuery()).Scan(&count); err != nil {
		return 0, fmt.Errorf("executing accountUserMemberships count query: %w", err)
	}

	return count, nil
}

// GetAllAccountUserMemberships fetches a list of all account user memberships in the database.
func (c *Client) GetAllAccountUserMemberships(ctx context.Context, results chan []*types.AccountUserMembership, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountUserMemberships called")

	count, countErr := c.GetAllAccountUserMembershipsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of accountUserMemberships: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfAccountUserMembershipsQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			accountUserMemberships, _, _, scanErr := c.scanAccountUserMemberships(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- accountUserMemberships
		}(beginID, endID)
	}

	return nil
}

// GetAccountUserMemberships fetches a list of account user memberships from the database that meet a particular filter.
func (c *Client) GetAccountUserMemberships(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.AccountUserMembershipList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountUserMembershipList{}

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetAccountUserMemberships called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAccountUserMembershipsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing accountUserMemberships list retrieval query: %w", err)
	}

	if x.AccountUserMemberships, x.FilteredCount, x.TotalCount, err = c.scanAccountUserMemberships(rows, true); err != nil {
		return nil, fmt.Errorf("scanning accountUserMemberships: %w", err)
	}

	return x, nil
}

// GetAccountUserMembershipsForAdmin fetches a list of account user memberships from the database that meet a particular filter for all users.
func (c *Client) GetAccountUserMembershipsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.AccountUserMembershipList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountUserMembershipList{}

	c.logger.Debug("GetAccountUserMembershipsForAdmin called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAccountUserMembershipsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing accountUserMemberships list retrieval query for admin: %w", err)
	}

	if x.AccountUserMemberships, x.FilteredCount, x.TotalCount, err = c.scanAccountUserMemberships(rows, true); err != nil {
		return nil, fmt.Errorf("scanning accountUserMemberships: %w", err)
	}

	return x, nil
}

// CreateAccountUserMembership creates an account user membership in the database.
func (c *Client) CreateAccountUserMembership(ctx context.Context, input *types.AccountUserMembershipCreationInput) (*types.AccountUserMembership, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccountUserMembership called")

	query, args := c.sqlQueryBuilder.BuildCreateAccountUserMembershipQuery(input)

	// create the accountUserMembership.
	id, err := c.performCreateQuery(ctx, c.db, false, "accountUserMembership creation", query, args)
	if err != nil {
		return nil, err
	}

	x := &types.AccountUserMembership{
		ID: id,
		//
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	return x, nil
}

// ArchiveAccountUserMembership archives an account user membership from the database by its ID.
func (c *Client) ArchiveAccountUserMembership(ctx context.Context, accountUserMembershipID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountUserMembershipIDToSpan(span, accountUserMembershipID)

	c.logger.WithValues(map[string]interface{}{
		"accountUserMembership_id": accountUserMembershipID,
		keys.UserIDKey:             userID,
	}).Debug("ArchiveAccountUserMembership called")

	query, args := c.sqlQueryBuilder.BuildArchiveAccountUserMembershipQuery(accountUserMembershipID, userID)

	return c.performCreateQueryIgnoringReturn(ctx, c.db, "accountUserMembership archive", query, args)
}

// LogAccountUserMembershipCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountUserMembershipCreationEvent(ctx context.Context, accountUserMembership *types.AccountUserMembership) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, accountUserMembership.BelongsToUser).Debug("LogAccountUserMembershipCreationEvent called")

	c.createAuditLogEntry(ctx, c.db, audit.BuildAccountUserMembershipCreationEventEntry(accountUserMembership))
}

// LogAccountUserMembershipUpdateEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountUserMembershipUpdateEvent(ctx context.Context, userID, accountUserMembershipID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogAccountUserMembershipUpdateEvent called")

	c.createAuditLogEntry(ctx, c.db, audit.BuildAccountUserMembershipUpdateEventEntry(userID, accountUserMembershipID, changes))
}

// LogAccountUserMembershipArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountUserMembershipArchiveEvent(ctx context.Context, userID, accountUserMembershipID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogAccountUserMembershipArchiveEvent called")

	c.createAuditLogEntry(ctx, c.db, audit.BuildAccountUserMembershipArchiveEventEntry(userID, accountUserMembershipID))
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
