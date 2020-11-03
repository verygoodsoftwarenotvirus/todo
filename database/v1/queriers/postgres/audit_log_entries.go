package postgres

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
func (p *Postgres) scanAuditLogEntry(scan database.Scanner) (*models.AuditLogEntry, error) {
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
func (p *Postgres) scanAuditLogEntries(rows database.ResultIterator) ([]models.AuditLogEntry, error) {
	var (
		list []models.AuditLogEntry
	)

	for rows.Next() {
		x, err := p.scanAuditLogEntry(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		p.logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (p *Postgres) buildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): entryID,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (p *Postgres) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	query, args := p.buildGetAuditLogEntryQuery(entryID)
	row := p.db.QueryRowContext(ctx, query, args...)
	return p.scanAuditLogEntry(row)
}

// buildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (p *Postgres) buildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, auditLogEntriesTableName)).
		From(auditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, archivedOnColumn): nil,
		}).
		ToSql()
	p.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// GetAllAuditLogEntriesCount will fetch the count of  from the database.
func (p *Postgres) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllAuditLogEntriesCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (p *Postgres) buildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := p.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn): endID,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetAllAuditLogEntries fetches every audit log entry from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (p *Postgres) GetAllAuditLogEntries(ctx context.Context, resultChannel chan []models.AuditLogEntry) error {
	count, err := p.GetAllAuditLogEntriesCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := p.buildGetBatchOfAuditLogEntriesQuery(begin, end)
			logger := p.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := p.db.Query(query, args...)
			if err == sql.ErrNoRows {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			auditLogEntries, err := p.scanAuditLogEntries(rows)
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
func (p *Postgres) buildGetAuditLogEntriesQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
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
	p.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntries fetches a list of  from the database that meet a particular filter.
func (p *Postgres) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	query, args := p.buildGetAuditLogEntriesQuery(filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for ")
	}

	auditLogEntries, err := p.scanAuditLogEntries(rows)
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
func (p *Postgres) buildCreateAuditLogEntryQuery(input *models.AuditLogEntry) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Insert(auditLogEntriesTableName).
		Columns(
			auditLogEntriesTableEventTypeColumn,
			auditLogEntriesTableContextColumn,
		).
		Values(
			input.EventType,
			input.Context,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", idColumn, createdOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// CreateAuditLogEntry creates an audit log entry in the database.
func (p *Postgres) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	x := &models.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := p.buildCreateAuditLogEntryQuery(x)

	// create the audit log entry.
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		p.logger.WithValue("event_type", input.EventType).Error(err, "executing audit log entry creation query")
	}
}

// createAuditLogEntry creates an audit log entry in the database.
func (p *Postgres) createAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	x := &models.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := p.buildCreateAuditLogEntryQuery(x)

	// create the audit log entry.
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		p.logger.WithValue("event_type", input.EventType).Error(err, "executing audit log entry creation query")
	}
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (p *Postgres) LogItemCreationEvent(ctx context.Context, item *models.Item) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemCreationEvent,
		Context: map[string]interface{}{
			"created": item,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (p *Postgres) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []models.FieldChangeSummary) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemUpdateEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"item_id":      itemID,
			"changes":      changes,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (p *Postgres) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.ItemArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"item_id":      itemID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (p *Postgres) LogOAuth2ClientCreationEvent(ctx context.Context, client *models.OAuth2Client) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.OAuth2ClientCreationEvent,
		Context: map[string]interface{}{
			"client": client,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (p *Postgres) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.OAuth2ClientArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"client_id":    clientID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (p *Postgres) LogUserCreationEvent(ctx context.Context, user *models.User) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserCreationEvent,
		Context: map[string]interface{}{
			"user": user,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (p *Postgres) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserVerifyTwoFactorSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (p *Postgres) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserUpdateTwoFactorSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (p *Postgres) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserUpdatePasswordEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (p *Postgres) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UserArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (p *Postgres) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.CycleCookieSecretEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (p *Postgres) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.SuccessfulLoginEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (p *Postgres) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UnsuccessfulLoginBadPasswordEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (p *Postgres) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.UnsuccessfulLoginBad2FATokenEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (p *Postgres) LogLogoutEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.LogoutEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (p *Postgres) LogWebhookCreationEvent(ctx context.Context, webhook *models.Webhook) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookCreationEvent,
		Context: map[string]interface{}{
			"webhook": webhook,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (p *Postgres) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []models.FieldChangeSummary) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookUpdateEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"webhook_id":   webhookID,
			"changes":      changes,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (p *Postgres) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.WebhookArchiveEvent,
		Context: map[string]interface{}{
			"performed_by": userID,
			"webhook_id":   webhookID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}
