package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	auditLogEntriesTableName            = "audit_log"
	auditLogEntriesTableEventTypeColumn = "event_type"
	auditLogEntriesTableContextColumn   = "context"
)

var (
	auditLogEntriesTableColumns = []string{
		fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn),
		fmt.Sprintf("%s.%s", auditLogEntriesTableName, auditLogEntriesTableEventTypeColumn),
		fmt.Sprintf("%s.%s", auditLogEntriesTableName, auditLogEntriesTableContextColumn),
		fmt.Sprintf("%s.%s", auditLogEntriesTableName, createdOnColumn),
	}
)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct
func (s *Sqlite) scanAuditLogEntry(scan database.Scanner) (*models.AuditLogEntry, error) {
	x := &models.AuditLogEntry{}

	targetVars := []interface{}{
		&x.ID,
		&x.EventType,
		&x.Context,
		&x.CreatedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanAuditLogEntries takes a logger and some database rows and turns them into a slice of .
func (s *Sqlite) scanAuditLogEntries(rows database.ResultIterator) ([]models.AuditLogEntry, error) {
	var (
		list []models.AuditLogEntry
	)

	for rows.Next() {
		x, err := s.scanAuditLogEntry(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		s.logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): entryID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (s *Sqlite) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	query, args := s.buildGetAuditLogEntryQuery(entryID)
	row := s.db.QueryRowContext(ctx, query, args...)
	return s.scanAuditLogEntry(row)
}

// buildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (s *Sqlite) buildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, auditLogEntriesTableName)).
		From(auditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, archivedOnColumn): nil,
		}).
		ToSql()
	s.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// GetAllAuditLogEntriesCount will fetch the count of  from the database.
func (s *Sqlite) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	err = s.db.QueryRowContext(ctx, s.buildGetAllAuditLogEntriesCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (s *Sqlite) buildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := s.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): endID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetAllAuditLogEntries fetches every audit log entry from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (s *Sqlite) GetAllAuditLogEntries(ctx context.Context, resultChannel chan []models.AuditLogEntry) error {
	count, err := s.GetAllAuditLogEntriesCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := s.buildGetBatchOfAuditLogEntriesQuery(begin, end)
			logger := s.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := s.db.Query(query, args...)
			if err == sql.ErrNoRows {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			auditLogEntries, err := s.scanAuditLogEntries(rows)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- auditLogEntries
		}(beginID, endID)
	}

	return nil
}

// buildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (s *Sqlite) buildGetAuditLogEntriesQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, archivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, auditLogEntriesTableName)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntries fetches a list of  from the database that meet a particular filter.
func (s *Sqlite) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	query, args := s.buildGetAuditLogEntriesQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for ")
	}

	auditLogEntries, err := s.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &models.AuditLogEntryList{
		Pagination: models.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Entries: auditLogEntries,
	}

	return list, nil
}

// buildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (s *Sqlite) buildCreateAuditLogEntryQuery(input *models.AuditLogEntry) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Insert(auditLogEntriesTableName).
		Columns(
			auditLogEntriesTableEventTypeColumn,
			auditLogEntriesTableContextColumn,
		).
		Values(
			input.EventType,
			input.Context,
		).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// CreateAuditLogEntry creates an audit log entry in the database.
func (s *Sqlite) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	x := &models.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := s.buildCreateAuditLogEntryQuery(x)

	// create the audit log entry.
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		s.logger.WithValue("event_type", input.EventType).Error(err, "executing audit log entry creation query")
	}
}

// createAuditLogEntry creates an audit log entry in the database.
func (s *Sqlite) createAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	x := &models.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := s.buildCreateAuditLogEntryQuery(x)

	// create the audit log entry.
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		s.logger.WithValue("event_type", input.EventType).Error(err, "executing audit log entry creation query")
	}
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (s *Sqlite) LogItemCreationEvent(ctx context.Context, item *models.Item) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemCreationEvent,
		Context: map[string]interface{}{
			"created": item,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (s *Sqlite) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []models.FieldChangeEvent) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemUpdateEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"item_id":      itemID,
			"changes":      changes,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (s *Sqlite) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"item_id":      itemID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (s *Sqlite) LogOAuth2ClientCreationEvent(ctx context.Context, client *models.OAuth2Client) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.OAuth2ClientCreationEvent,
		Context: map[string]interface{}{
			"client": client,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (s *Sqlite) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.OAuth2ClientArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"client_id":    clientID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (s *Sqlite) LogUserCreationEvent(ctx context.Context, user *models.User) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserCreationEvent,
		Context: map[string]interface{}{
			"user": user,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (s *Sqlite) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserVerifyTwoFactorSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (s *Sqlite) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserUpdateTwoFactorSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (s *Sqlite) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserUpdatePasswordEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (s *Sqlite) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (s *Sqlite) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.CycleCookieSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (s *Sqlite) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.SuccessfulLoginEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (s *Sqlite) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UnsuccessfulLoginBadPasswordEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (s *Sqlite) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UnsuccessfulLoginBad2FATokenEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (s *Sqlite) LogLogoutEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.LogoutEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (s *Sqlite) LogWebhookCreationEvent(ctx context.Context, webhook *models.Webhook) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookCreationEvent,
		Context: map[string]interface{}{
			"webhook": webhook,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (s *Sqlite) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []models.FieldChangeEvent) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookUpdateEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"webhook_id":   webhookID,
			"changes":      changes,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (s *Sqlite) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"webhook_id":   webhookID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}
