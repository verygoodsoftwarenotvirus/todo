package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// ExportData extracts all the data for a given user and puts it in a fat ol' struct for export/import use
func (p *Postgres) ExportData(ctx context.Context, user *models.User) (*models.DataExport, error) {
	x := models.NewDataExport()
	x.User = *user

	il, err := p.GetAllItemsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	x.Items = il

	ol, err := p.GetAllOAuth2ClientsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for _, o := range ol {
		x.OAuth2Clients = append(x.OAuth2Clients, *o)
	}

	wl, err := p.GetAllWebhooksForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	x.Webhooks = wl

	return x, nil
}
