package client

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	auditLogBasePath = "/_admin_/audit_log"
)

// BuildGetAuditLogEntriesRequest builds an HTTP request for fetching entrys.
func (c *V1Client) BuildGetAuditLogEntriesRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildGetAuditLogEntriesRequest")
	defer span.End()

	uri := c.buildVersionlessURL(
		filter.ToValues(),
		auditLogBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogEntries retrieves a list of entrys.
func (c *V1Client) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (entrys *models.AuditLogEntryList, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntries")
	defer span.End()

	req, err := c.BuildGetAuditLogEntriesRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entries")
	if retrieveErr := c.retrieve(ctx, req, &entrys); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entrys, nil
}
