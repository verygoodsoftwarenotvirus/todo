package client

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const oauth2ClientsBasePath = "oauth2/clients"

// GetOAuth2Client gets an OAuth2 client
func (c *V1Client) GetOAuth2Client(ctx context.Context, id uint64) (oauth2Client *models.OAuth2Client, err error) {
	uri := c.BuildURL(nil, oauth2ClientsBasePath, strconv.FormatUint(id, 10))
	return oauth2Client, c.get(ctx, uri, &oauth2Client)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *V1Client) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter) (oauth2Clients *models.OAuth2ClientList, err error) {
	uri := c.BuildURL(filter.ToValues(), oauth2ClientsBasePath)
	return oauth2Clients, c.get(ctx, uri, &oauth2Clients)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *V1Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput, cookie *http.Cookie) (oauth2Client *models.OAuth2Client, err error) {
	if cookie == nil && c.currentUserCookie == nil {
		return nil, errors.New("no cookie available for authenticated request")
	} else if cookie == nil {
		cookie = c.currentUserCookie
	}

	uri := c.buildVersionlessURL(nil, "oauth2", "client")
	// I can ignore this error because I know that URI will be valid
	req, err := c.buildDataRequest(http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}
	req.AddCookie(cookie)

	res, err := c.executeRequest(ctx, c.plainClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error executing request")
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	resErr := unmarshalBody(res, &oauth2Client)
	if resErr != nil {
		return nil, errors.Wrap(err, "encountered error loading response from server")
	}

	return
}

// UpdateOAuth2Client updates an OAuth2 client
func (c *V1Client) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	logger := c.logger.WithValues(map[string]interface{}{
		"id":        updated.ID,
		"client_id": updated.ClientID,
	})

	ctx, span := trace.StartSpan(ctx, "UpdateOAuth2Client")
	defer span.End()
	idStr := strconv.FormatUint(updated.ID, 10)
	span.AddAttributes(trace.StringAttribute("oauth2_client_id", idStr))

	uri := c.BuildURL(nil, oauth2ClientsBasePath, idStr)
	if err := c.put(ctx, uri, updated, &updated); err != nil {
		logger.Error(err, "error encountered updating OAuth2 client")
		return err
	}
	return nil
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *V1Client) DeleteOAuth2Client(ctx context.Context, id uint64) error {
	logger := c.logger.WithValue("oauth2client_id", id)

	ctx, span := trace.StartSpan(ctx, "DeleteOAuth2Client")
	defer span.End()
	idStr := strconv.FormatUint(id, 10)
	span.AddAttributes(trace.StringAttribute("oauth2_client_id", idStr))

	uri := c.BuildURL(nil, oauth2ClientsBasePath, strconv.FormatUint(id, 10))
	if err := c.delete(ctx, uri); err != nil {
		logger.Error(err, "error encountered deleting OAuth2 client")
		return err
	}
	return nil
}
