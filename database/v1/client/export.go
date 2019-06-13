package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

// ExportData extracts all the data for a given user and puts it in a fat ol' struct for export/import use
func (c *Client) ExportData(ctx context.Context, user *models.User) (*models.DataExport, error) {
	ctx, span := trace.StartSpan(ctx, "ExportData")
	defer span.End()

	c.logger.Debug("ExportData called")

	return c.querier.ExportData(ctx, user)
}
