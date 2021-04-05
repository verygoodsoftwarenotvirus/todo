package querier

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.UserDataManager = (*SQLQuerier)(nil)
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (q *SQLQuerier) scanUser(ctx context.Context, scan database.Scanner, includeCounts bool) (user *types.User, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)
	user = &types.User{}

	var perms uint32

	targetVars := []interface{}{
		&user.ID,
		&user.ExternalID,
		&user.Username,
		&user.AvatarSrc,
		&user.HashedPassword,
		&user.Salt,
		&user.RequiresPasswordChange,
		&user.PasswordLastChangedOn,
		&user.TwoFactorSecret,
		&user.TwoFactorSecretVerifiedOn,
		&perms,
		&user.Reputation,
		&user.ReputationExplanation,
		&user.CreatedOn,
		&user.LastUpdatedOn,
		&user.ArchivedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "scanning user")
	}

	user.ServiceAdminPermissions = permissions.NewServiceAdminPermissions(perms)

	return user, filteredCount, totalCount, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (q *SQLQuerier) scanUsers(ctx context.Context, rows database.ResultIterator, includeCounts bool) (users []*types.User, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		user, fc, tc, scanErr := q.scanUser(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, observability.PrepareError(scanErr, logger, span, "scanning user result")
		}

		if includeCounts && filteredCount == 0 {
			filteredCount = fc
		}

		if includeCounts && totalCount == 0 {
			totalCount = tc
		}

		users = append(users, user)
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return users, filteredCount, totalCount, nil
}

// getUser fetches a user.
func (q *SQLQuerier) getUser(ctx context.Context, userID uint64, withVerifiedTOTPSecret bool) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	var (
		query string
		args  []interface{}
	)

	if withVerifiedTOTPSecret {
		query, args = q.sqlQueryBuilder.BuildGetUserQuery(ctx, userID)
	} else {
		query, args = q.sqlQueryBuilder.BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx, userID)
	}

	row := q.getOneRow(ctx, "user", query, args...)

	u, _, _, err := q.scanUser(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

// UserHasStatus fetches whether or not an item exists from the database.
func (q *SQLQuerier) UserHasStatus(ctx context.Context, userID uint64, statuses ...string) (banned bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return false, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID).WithValue("statuses", statuses)
	tracing.AttachUserIDToSpan(span, userID)

	query, args := q.sqlQueryBuilder.BuildUserHasStatusQuery(ctx, userID, statuses...)

	result, err := q.performBooleanQuery(ctx, q.db, query, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing user status check")
	}

	return result, nil
}

// GetUser fetches a user.
func (q *SQLQuerier) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)

	return q.getUser(ctx, userID, true)
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (q *SQLQuerier) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)

	return q.getUser(ctx, userID, false)
}

// GetUserByUsername fetches a user by their username.
func (q *SQLQuerier) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, ErrEmptyInputProvided
	}

	tracing.AttachUsernameToSpan(span, username)
	logger := q.logger.WithValue(keys.UsernameKey, username)

	query, args := q.sqlQueryBuilder.BuildGetUserByUsernameQuery(ctx, username)
	row := q.getOneRow(ctx, "user", query, args...)

	u, _, _, err := q.scanUser(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (q *SQLQuerier) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if usernameQuery == "" {
		return []*types.User{}, ErrEmptyInputProvided
	}

	tracing.AttachSearchQueryToSpan(span, usernameQuery)
	logger := q.logger.WithValue(keys.SearchQueryKey, usernameQuery)

	query, args := q.sqlQueryBuilder.BuildSearchForUserByUsernameQuery(ctx, usernameQuery)

	rows, err := q.performReadQuery(ctx, "user search by username", query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying database for users")
	}

	u, _, _, err := q.scanUsers(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

// GetAllUsersCount fetches a count of users from the database that meet a particular filter.
func (q *SQLQuerier) GetAllUsersCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllUsersCountQuery(ctx), "fetching count of users")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of users")
	}

	return count, nil
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (q *SQLQuerier) GetUsers(ctx context.Context, filter *types.QueryFilter) (x *types.UserList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.UserList{}

	tracing.AttachQueryFilterToSpan(span, filter)
	logger := filter.AttachToLogger(q.logger)

	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetUsersQuery(ctx, filter)

	rows, err := q.performReadQuery(ctx, "users", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	if x.Users, x.FilteredCount, x.TotalCount, err = q.scanUsers(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "loading response from database")
	}

	return x, nil
}

// createUser creates a user. The `user` and `account` parameters are meant to be filled out.
func (q *SQLQuerier) createUser(ctx context.Context, user *types.User, account *types.Account, userCreationQuery string, userCreationArgs []interface{}) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("username", user.Username)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	userID, err := q.performWriteQuery(ctx, tx, false, "user creation", userCreationQuery, userCreationArgs)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "creating user")
	}

	user.ID = userID
	account.BelongsToUser = user.ID

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserCreationEventEntry(user.ID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user creation audit log entry")
	}

	// create the account.
	accountCreationInput := types.NewAccountCreationInputForUser(user)
	accountCreationInput.DefaultUserPermissions = account.DefaultUserPermissions
	accountCreationQuery, accountCreationArgs := q.sqlQueryBuilder.BuildAccountCreationQuery(ctx, accountCreationInput)

	accountID, err := q.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "create account")
	}

	account.ID = accountID

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account, user.ID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account creation audit log entry")
	}

	addUserToAccountQuery, addUserToAccountArgs := q.sqlQueryBuilder.BuildCreateMembershipForNewUserQuery(ctx, userID, accountID)
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account user membership creation", addUserToAccountQuery, addUserToAccountArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account user membership creation audit log entry")
	}

	addToAccountInput := &types.AddUserToAccountInput{
		UserID:                 user.ID,
		UserAccountPermissions: account.DefaultUserPermissions,
		Reason:                 "account creation",
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(userID, account.ID, addToAccountInput)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user added to account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachAccountIDToSpan(span, account.ID)

	return nil
}

// CreateUser creates a user.
func (q *SQLQuerier) CreateUser(ctx context.Context, input *types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachUsernameToSpan(span, input.Username)
	logger := q.logger.WithValue(keys.UsernameKey, input.Username)

	// create the user.
	userCreationQuery, userCreationArgs := q.sqlQueryBuilder.BuildCreateUserQuery(ctx, input)

	user := &types.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
		CreatedOn:       q.currentTime(),
	}

	account := &types.Account{
		Name:                   input.Username,
		PlanID:                 nil,
		CreatedOn:              q.currentTime(),
		DefaultUserPermissions: permissions.ServiceUserPermissions(math.MaxUint32),
	}

	if err := q.createUser(ctx, user, account, userCreationQuery, userCreationArgs); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating user")
	}

	return user, nil
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (q *SQLQuerier) UpdateUser(ctx context.Context, updated *types.User, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	tracing.AttachUsernameToSpan(span, updated.Username)
	logger := q.logger.WithValue(keys.UsernameKey, updated.Username)

	query, args := q.sqlQueryBuilder.BuildUpdateUserQuery(ctx, updated)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating user")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateEventEntry(updated.ID, changes)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// UpdateUserPassword updates a user's authentication hash in the database.
func (q *SQLQuerier) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	if newHash == "" {
		return ErrEmptyInputProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	query, args := q.sqlQueryBuilder.BuildUpdateUserPasswordQuery(ctx, userID, newHash)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user authentication update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating user password")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdatePasswordEventEntry(userID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user password update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// UpdateUserTwoFactorSecret marks a user's two factor secret as validated.
func (q *SQLQuerier) UpdateUserTwoFactorSecret(ctx context.Context, userID uint64, newSecret string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	if newSecret == "" {
		return ErrEmptyInputProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	query, args := q.sqlQueryBuilder.BuildUpdateUserTwoFactorSecretQuery(ctx, userID, newSecret)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user 2FA secret update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating user 2FA secret")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing update 2FA secret audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (q *SQLQuerier) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	query, args := q.sqlQueryBuilder.BuildVerifyUserTwoFactorSecretQuery(ctx, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret verification", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing verified two factor status to database")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user 2FA secret verification audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// ArchiveUser archives a user.
func (q *SQLQuerier) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	archiveUserQuery, archiveUserArgs := q.sqlQueryBuilder.BuildArchiveUserQuery(ctx, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user archive", archiveUserQuery, archiveUserArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserArchiveEventEntry(userID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user archive audit log entry")
	}

	archiveMembershipsQuery, archiveMembershipsArgs := q.sqlQueryBuilder.BuildArchiveAccountMembershipsForUserQuery(ctx, userID)

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user memberships archive", archiveMembershipsQuery, archiveMembershipsArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user account memberships")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that relate to a given user.
func (q *SQLQuerier) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesForUserQuery(ctx, userID)

	rows, err := q.performReadQuery(ctx, "audit log entries for user", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}
