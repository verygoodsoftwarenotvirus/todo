package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	itemsBasePath = "items"
)

// BuildItemExistsRequest builds an HTTP request for checking the existence of an item.
func (c *Client) BuildItemExistsRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
}

// BuildGetItemRequest builds an HTTP request for fetching an item.
func (c *Client) BuildGetItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildSearchItemsRequest builds an HTTP request for querying items.
func (c *Client) BuildSearchItemsRequest(ctx context.Context, query string, limit uint8) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	params := url.Values{}
	params.Set(types.SearchQueryKey, query)
	params.Set(types.LimitQueryKey, strconv.FormatUint(uint64(limit), 10))

	uri := c.BuildURL(
		ctx,
		params,
		itemsBasePath,
		"search",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetItemsRequest builds an HTTP request for fetching items.
func (c *Client) BuildGetItemsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), itemsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateItemRequest builds an HTTP request for creating an item.
func (c *Client) BuildCreateItemRequest(ctx context.Context, input *types.ItemCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, itemsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateItemRequest builds an HTTP request for updating an item.
func (c *Client) BuildUpdateItemRequest(ctx context.Context, item *types.Item) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if item == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		strconv.FormatUint(item.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPut, uri, item)
}

// BuildArchiveItemRequest builds an HTTP request for updating an item.
func (c *Client) BuildArchiveItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForItemRequest builds an HTTP request for fetching a list of audit log entries pertaining to an item.
func (c *Client) BuildGetAuditLogForItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, itemsBasePath, strconv.FormatUint(itemID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
