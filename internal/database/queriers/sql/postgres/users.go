package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

var (
	_ types.UserDataManager = (*SQLQuerier)(nil)
)

const (
	serviceRolesSeparator = ","
)

// scanUser provides a consistent way to scan something like a *sql.Row into a Requester struct.
func (q *SQLQuerier) scanUser(ctx context.Context, scan database.Scanner, includeCounts bool) (user *types.User, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)
	user = &types.User{}
	var rawRoles string

	targetVars := []interface{}{
		&user.ID,
		&user.Username,
		&user.AvatarSrc,
		&user.HashedPassword,
		&user.RequiresPasswordChange,
		&user.PasswordLastChangedOn,
		&user.TwoFactorSecret,
		&user.TwoFactorSecretVerifiedOn,
		&rawRoles,
		&user.ServiceAccountStatus,
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

	if roles := strings.Split(rawRoles, serviceRolesSeparator); len(roles) > 0 {
		user.ServiceRoles = roles
	} else {
		user.ServiceRoles = []string{}
	}

	return user, filteredCount, totalCount, nil
}

// scanUsers takes database rows and loads them into a slice of Requester structs.
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

const getUserQuery = `
	SELECT
		users.id,
		users.username,
		users.avatar_src,
		users.hashed_password,
		users.requires_password_change,
		users.password_last_changed_on,
		users.two_factor_secret,
		users.two_factor_secret_verified_on,
		users.service_roles,
		users.reputation,
		users.reputation_explanation,
		users.created_on,
		users.last_updated_on,
		users.archived_on
	FROM users
	WHERE users.archived_on IS NULL
	AND users.id = $1
	AND users.two_factor_secret_verified_on IS NOT NULL
`

const getUserWithUnverifiedTwoFactorSecretQuery = `
	SELECT 
		users.id, 
		users.username, 
		users.avatar_src, 
		users.hashed_password, 
		users.requires_password_change, 
		users.password_last_changed_on, 
		users.two_factor_secret, 
		users.two_factor_secret_verified_on, 
		users.service_roles, 
		users.reputation, 
		users.reputation_explanation, 
		users.created_on, 
		users.last_updated_on, 
		users.archived_on 
	FROM users 
	WHERE users.archived_on IS NULL 
	AND users.id = $1
	AND users.two_factor_secret_verified_on IS NULL
`

// getUser fetches a user.
func (q *SQLQuerier) getUser(ctx context.Context, userID string, withVerifiedTOTPSecret bool) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	var query string
	args := []interface{}{userID}

	if withVerifiedTOTPSecret {
		query = getUserQuery
	} else {
		query = getUserWithUnverifiedTwoFactorSecretQuery
	}

	row := q.getOneRow(ctx, q.db, "user", query, args)

	u, _, _, err := q.scanUser(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

const createAccountMembershipForNewUserQuery = `
	INSERT INTO account_user_memberships (id,belongs_to_user,belongs_to_account,default_account,account_roles) 
	VALUES ($1,$2,$3,$4,$5)
`

// createUser creates a user. The `user` and `account` parameters are meant to be filled out.
func (q *SQLQuerier) createUser(ctx context.Context, user *types.User, account *types.Account, userCreationQuery string, userCreationArgs []interface{}) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("username", user.Username)

	if user.ID == "" {
		return ErrEmptyInputProvided
	}
	account.BelongsToUser = user.ID
	logger = logger.WithValue(keys.UserIDKey, user.ID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if writeErr := q.performWriteQuery(ctx, tx, "user creation", userCreationQuery, userCreationArgs); writeErr != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(writeErr, logger, span, "creating user")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserCreationEventEntry(user.ID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user creation audit log entry")
	}

	// create the account.
	accountCreationInput := types.AccountCreationInputForNewUser(user)
	accountCreationInput.ID = account.ID

	accountCreationArgs := []interface{}{
		accountCreationInput.ID,
		accountCreationInput.Name,
		types.UnpaidAccountBillingStatus,
		accountCreationInput.ContactEmail,
		accountCreationInput.ContactPhone,
		accountCreationInput.BelongsToUser,
	}

	if writeErr := q.performWriteQuery(ctx, tx, "account creation", accountCreationQuery, accountCreationArgs); writeErr != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(writeErr, logger, span, "create account")
	}

	logger = logger.WithValue(keys.AccountIDKey, account.ID)

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account, user.ID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account creation audit log entry")
	}

	createAccountMembershipForNewUserArgs := []interface{}{
		ksuid.New().String(),
		user.ID,
		account.ID,
		true,
		authorization.AccountAdminRole.String(),
	}

	if err = q.performWriteQuery(ctx, tx, "account user membership creation", createAccountMembershipForNewUserQuery, createAccountMembershipForNewUserArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account user membership creation audit log entry")
	}

	addToAccountInput := &types.AddUserToAccountInput{
		UserID:       user.ID,
		AccountID:    account.ID,
		AccountRoles: []string{authorization.AccountMemberRole.String()},
		Reason:       "account creation",
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(user.ID, addToAccountInput)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user added to account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachAccountIDToSpan(span, account.ID)

	logger.Info("user and account created")

	return nil
}

const userHasStatusQuery = `
	SELECT EXISTS ( SELECT users.id FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND (users.reputation = $2 OR users.reputation = $3) )
`

// UserHasStatus fetches whether an user has a particular status.
func (q *SQLQuerier) UserHasStatus(ctx context.Context, userID string, statuses ...string) (banned bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return false, ErrInvalidIDProvided
	}

	if len(statuses) == 0 {
		return true, nil
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID).WithValue("statuses", statuses)
	tracing.AttachUserIDToSpan(span, userID)

	args := []interface{}{userID}
	for _, status := range statuses {
		args = append(args, status)
	}

	result, err := q.performBooleanQuery(ctx, q.db, userHasStatusQuery, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing user status check")
	}

	return result, nil
}

// GetUser fetches a user.
func (q *SQLQuerier) GetUser(ctx context.Context, userID string) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)

	return q.getUser(ctx, userID, true)
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (q *SQLQuerier) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID string) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)

	return q.getUser(ctx, userID, false)
}

const getUserByUsernameQuery = `
	SELECT 
		users.id, 
		users.username, 
		users.avatar_src, 
		users.hashed_password, 
		users.requires_password_change, 
		users.password_last_changed_on, 
		users.two_factor_secret, 
		users.two_factor_secret_verified_on, 
		users.service_roles, 
		users.reputation, 
		users.reputation_explanation, 
		users.created_on, 
		users.last_updated_on, 
		users.archived_on 
	FROM users 
	WHERE users.archived_on IS NULL 
	AND users.username = $1
	AND users.two_factor_secret_verified_on IS NOT NULL
`

// GetUserByUsername fetches a user by their username.
func (q *SQLQuerier) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, ErrEmptyInputProvided
	}

	tracing.AttachUsernameToSpan(span, username)
	logger := q.logger.WithValue(keys.UsernameKey, username)

	args := []interface{}{username}

	row := q.getOneRow(ctx, q.db, "user", getUserByUsernameQuery, args)

	u, _, _, err := q.scanUser(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

const searchForUserByUsernameQuery = `
	SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.username ILIKE $1 AND users.archived_on IS NULL AND users.two_factor_secret_verified_on IS NOT NULL	
`

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (q *SQLQuerier) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if usernameQuery == "" {
		return []*types.User{}, ErrEmptyInputProvided
	}

	tracing.AttachSearchQueryToSpan(span, usernameQuery)
	logger := q.logger.WithValue(keys.SearchQueryKey, usernameQuery)

	args := []interface{}{
		fmt.Sprintf("%s%%", usernameQuery),
	}

	rows, err := q.performReadQuery(ctx, q.db, "user search by username", searchForUserByUsernameQuery, args)
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

const getAllUsersCountQuery = `SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL`

// GetAllUsersCount fetches a count of users from the database that meet a particular filter.
func (q *SQLQuerier) GetAllUsersCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, getAllUsersCountQuery, "fetching count of users")
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

	query, args := q.buildListQuery(
		ctx,
		querybuilding.UsersTableName,
		nil,
		nil,
		"",
		querybuilding.UsersTableColumns,
		"",
		false,
		filter,
	)

	rows, err := q.performReadQuery(ctx, q.db, "users", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	if x.Users, x.FilteredCount, x.TotalCount, err = q.scanUsers(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "loading response from database")
	}

	return x, nil
}

const userCreationQuery = `
	INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles) VALUES ($1,$2,$3,$4,$5,$6)
`

// CreateUser creates a user.
func (q *SQLQuerier) CreateUser(ctx context.Context, input *types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachUsernameToSpan(span, input.Username)
	logger := q.logger.WithValue(keys.UsernameKey, input.Username)

	userCreationArgs := []interface{}{
		input.ID,
		input.Username,
		input.HashedPassword,
		input.TwoFactorSecret,
		types.UnverifiedAccountStatus,
		authorization.ServiceUserRole.String(),
	}

	user := &types.User{
		ID:              input.ID,
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
		ServiceRoles:    []string{authorization.ServiceUserRole.String()},
		CreatedOn:       q.currentTime(),
	}

	account := &types.Account{
		ID:                 ksuid.New().String(),
		Name:               input.Username,
		SubscriptionPlanID: nil,
		CreatedOn:          q.currentTime(),
	}

	// create the user.
	if err := q.createUser(ctx, user, account, userCreationQuery, userCreationArgs); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating user")
	}

	return user, nil
}

const updateUserQuery = `
	UPDATE users SET username = $1, hashed_password = $2, avatar_src = $3, two_factor_secret = $4, two_factor_secret_verified_on = $5, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $6
`

// UpdateUser receives a complete Requester struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (q *SQLQuerier) UpdateUser(ctx context.Context, updated *types.User, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	tracing.AttachUsernameToSpan(span, updated.Username)
	logger := q.logger.WithValue(keys.UsernameKey, updated.Username)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	args := []interface{}{
		updated.Username,
		updated.HashedPassword,
		updated.AvatarSrc,
		updated.TwoFactorSecret,
		updated.TwoFactorSecretVerifiedOn,
		updated.ID,
	}

	if err = q.performWriteQuery(ctx, tx, "user update", updateUserQuery, args); err != nil {
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

	logger.Info("user updated")

	return nil
}

// UpdateUserPassword updates a user's passwords hash in the database.
func (q *SQLQuerier) UpdateUserPassword(ctx context.Context, userID, newHash string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if newHash == "" {
		return ErrEmptyInputProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query := `
		UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $3
	`
	args := []interface{}{
		newHash,
		false,
		userID,
	}

	if err = q.performWriteQuery(ctx, tx, "user passwords update", query, args); err != nil {
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

	logger.Info("user password updated")

	return nil
}

// UpdateUserTwoFactorSecret marks a user's two factor secret as validated.
func (q *SQLQuerier) UpdateUserTwoFactorSecret(ctx context.Context, userID, newSecret string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if newSecret == "" {
		return ErrEmptyInputProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query := "UPDATE users SET two_factor_secret_verified_on = $1, two_factor_secret = $2 WHERE archived_on IS NULL AND id = $3"
	args := []interface{}{
		nil,
		newSecret,
		userID,
	}

	if err = q.performWriteQuery(ctx, tx, "user 2FA secret update", query, args); err != nil {
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

	logger.Info("user two factor secret updated")

	return nil
}

// MarkUserTwoFactorSecretAsVerified marks a user's two factor secret as validated.
func (q *SQLQuerier) MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()), reputation = $1 WHERE archived_on IS NULL AND id = $2"
	args := []interface{}{
		types.GoodStandingAccountStatus,
		userID,
	}

	if err = q.performWriteQuery(ctx, tx, "user two factor secret verification", query, args); err != nil {
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

	logger.Info("user two factor secret verified")

	return nil
}

const archiveUserQuery = `
	UPDATE users SET archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $1
`

// ArchiveUser archives a user.
func (q *SQLQuerier) ArchiveUser(ctx context.Context, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger := q.logger.WithValue(keys.UserIDKey, userID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	archiveUserArgs := []interface{}{userID}

	if err = q.performWriteQuery(ctx, tx, "user archive", archiveUserQuery, archiveUserArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user")
	}

	archiveMembershipsQuery := `
		UPDATE account_user_memberships SET archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1
	`
	archiveMembershipsArgs := []interface{}{userID}

	if err = q.performWriteQuery(ctx, tx, "user memberships archive", archiveMembershipsQuery, archiveMembershipsArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user account memberships")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserArchiveEventEntry(userID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("user archived")

	return nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that relate to a given user.
func (q *SQLQuerier) GetAuditLogEntriesForUser(ctx context.Context, userID string) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)

	query := `
		SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE (audit_log.context->>'user_id' = $1 OR audit_log.context->>'performed_by' = $2) ORDER BY audit_log.created_on
	`
	args := []interface{}{
		userID,
		userID,
	}

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries for user", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}
