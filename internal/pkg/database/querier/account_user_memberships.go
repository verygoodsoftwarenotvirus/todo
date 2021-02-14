package querier

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*Client)(nil)
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (c *Client) scanAccountUserMembership(scan database.Scanner) (x *types.AccountUserMembership, err error) {
	x = &types.AccountUserMembership{}

	targetVars := []interface{}{
		&x.ID,
		&x.BelongsToAccount,
		&x.BelongsToUser,
		&x.UserPermissions,
		&x.DefaultAccount,
		&x.CreatedOn,
		&x.ArchivedOn,
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, scanErr
	}

	return x, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of memberships.
func (c *Client) scanAccountUserMemberships(rows database.ResultIterator) (memberships []*types.AccountUserMembership, err error) {
	for rows.Next() {
		x, scanErr := c.scanAccountUserMembership(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		memberships = append(memberships, x)
	}

	if handleErr := c.handleRows(rows); handleErr != nil {
		return nil, handleErr
	}

	return memberships, nil
}

// GetMembershipsForUser does a thing.
func (c *Client) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]bitmask.ServiceUserPermissions, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	})

	logger.Debug("GetMembershipsForUser called")

	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	rows, getMembershipsErr := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if getMembershipsErr != nil {
		logger.WithValue("query", getAccountMembershipsQuery).Info("FUCK")
		return 0, nil, getMembershipsErr
	}

	memberships, scanErr := c.scanAccountUserMemberships(rows)
	if scanErr != nil {
		logger.WithValue("query", getAccountMembershipsQuery).Info("FUCK")
		return 0, nil, scanErr
	}

	permissionsMap = map[uint64]bitmask.ServiceUserPermissions{}

	for _, membership := range memberships {
		permissionsMap[membership.ID] = membership.UserPermissions

		if membership.DefaultAccount && defaultAccount == 0 {
			defaultAccount = membership.ID
		}
	}

	return defaultAccount, permissionsMap, nil
}
