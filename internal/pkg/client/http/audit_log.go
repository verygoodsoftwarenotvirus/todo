package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	adminBasePath    = "_admin_"
	auditLogBasePath = "audit_log"
)

// BuildGetAuditLogEntriesRequest builds an HTTP request for fetching entries.
func (c *V1Client) BuildGetAuditLogEntriesRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildGetAuditLogEntriesRequest")
	defer span.End()

	uri := c.buildURL(
		filter.ToValues(),
		adminBasePath,
		auditLogBasePath,
	).String()
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogEntries retrieves a list of entries.
func (c *V1Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (entries *types.AuditLogEntryList, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntries")
	defer span.End()

	req, err := c.BuildGetAuditLogEntriesRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entries")
	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

// BuildGetAuditLogEntryRequest builds an HTTP request for fetching entries.
func (c *V1Client) BuildGetAuditLogEntryRequest(ctx context.Context, entryID uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildGetAuditLogEntryRequest")
	defer span.End()

	uri := c.buildURL(
		nil,
		adminBasePath,
		auditLogBasePath,
		strconv.FormatUint(entryID, 10),
	).String()
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogEntry retrieves an entry.
func (c *V1Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (entry *types.AuditLogEntry, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntry")
	defer span.End()

	req, err := c.BuildGetAuditLogEntryRequest(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entry")
	if retrieveErr := c.retrieve(ctx, req, &entry); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entry, nil
}
