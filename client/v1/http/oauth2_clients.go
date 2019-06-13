package client

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	oauth2ClientsBasePath = "oauth2/clients"
)

// BuildGetOAuth2ClientRequest builds an http Request for fetching an oauth2 client
func (c *V1Client) BuildGetOAuth2ClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, oauth2ClientsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetOAuth2Client gets an OAuth2 client
func (c *V1Client) GetOAuth2Client(ctx context.Context, id uint64) (oauth2Client *models.OAuth2Client, err error) {
	logger := c.logger.WithValue("oauth2_client_id", id)
	logger.Debug("GetOAuth2Client called")

	req, err := c.BuildGetOAuth2ClientRequest(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &oauth2Client)
	return oauth2Client, err
}

// BuildGetOAuth2ClientsRequest builds an http Request for fetching a list of oauth2 clients
func (c *V1Client) BuildGetOAuth2ClientsRequest(
	ctx context.Context,
	filter *models.QueryFilter,
) (*http.Request, error) {
	uri := c.BuildURL(filter.ToValues(), oauth2ClientsBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *V1Client) GetOAuth2Clients(
	ctx context.Context,
	filter *models.QueryFilter,
) (oauth2Clients *models.OAuth2ClientList, err error) {
	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetOAuth2Clients called")

	req, err := c.BuildGetOAuth2ClientsRequest(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &oauth2Clients)
	return oauth2Clients, err
}

// BuildCreateOAuth2ClientRequest builds an http Request for creating oauth2 clients
func (c *V1Client) BuildCreateOAuth2ClientRequest(
	ctx context.Context,
	cookie *http.Cookie,
	body *models.OAuth2ClientCreationInput,
) (*http.Request, error) {
	uri := c.buildVersionlessURL(nil, "oauth2", "client")

	req, err := c.buildDataRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	req.AddCookie(cookie)

	return req, nil
}

// CreateOAuth2Client creates an OAuth2 client
func (c *V1Client) CreateOAuth2Client(
	ctx context.Context,
	cookie *http.Cookie,
	input *models.OAuth2ClientCreationInput,
) (oauth2Client *models.OAuth2Client, err error) {
	if cookie == nil {
		return nil, errors.New("cookie required for request")
	}

	req, err := c.BuildCreateOAuth2ClientRequest(ctx, cookie, input)
	if err != nil {
		return nil, err
	}

	res, err := c.executeRawRequest(ctx, c.plainClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resErr := unmarshalBody(res, &oauth2Client); resErr != nil {
		return nil, errors.Wrap(resErr, "loading response from server")
	}

	return oauth2Client, nil
}

// BuildDeleteOAuth2ClientRequest builds an http Request for updating oauth2 clients
func (c *V1Client) BuildDeleteOAuth2ClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, oauth2ClientsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *V1Client) DeleteOAuth2Client(ctx context.Context, id uint64) error {
	c.logger.WithValue("oauth2client_id", id).Debug("DeleteOAuth2Client called")

	req, err := c.BuildDeleteOAuth2ClientRequest(ctx, id)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, nil)
}
