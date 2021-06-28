package requests

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	itemsBasePath = "items"
)

// BuildItemExistsRequest builds an HTTP request for checking the existence of an item.
func (b *Builder) BuildItemExistsRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		id(itemID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetItemRequest builds an HTTP request for fetching an item.
func (b *Builder) BuildGetItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		id(itemID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildSearchItemsRequest builds an HTTP request for querying items.
func (b *Builder) BuildSearchItemsRequest(ctx context.Context, query string, limit uint8) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger.WithValue(types.SearchQueryKey, query).WithValue(types.LimitQueryKey, limit)

	params := url.Values{}
	params.Set(types.SearchQueryKey, query)
	params.Set(types.LimitQueryKey, strconv.FormatUint(uint64(limit), 10))

	uri := b.BuildURL(
		ctx,
		params,
		itemsBasePath,
		"search",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetItemsRequest builds an HTTP request for fetching a list of items.
func (b *Builder) BuildGetItemsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(b.logger)

	uri := b.BuildURL(
		ctx,
		filter.ToValues(),
		itemsBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)
	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildCreateItemRequest builds an HTTP request for creating an item.
func (b *Builder) BuildCreateItemRequest(ctx context.Context, input *types.ItemCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building request")
	}

	return req, nil
}

// BuildUpdateItemRequest builds an HTTP request for updating an item.
func (b *Builder) BuildUpdateItemRequest(ctx context.Context, item *types.Item) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if item == nil {
		return nil, ErrNilInputProvided
	}

	logger = logger.WithValue(keys.ItemIDKey, item.ID)
	tracing.AttachItemIDToSpan(span, item.ID)

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		strconv.FormatUint(item.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := b.buildDataRequest(ctx, http.MethodPut, uri, item)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building request")
	}

	return req, nil
}

// BuildArchiveItemRequest builds an HTTP request for archiving an item.
func (b *Builder) BuildArchiveItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		id(itemID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetAuditLogForItemRequest builds an HTTP request for fetching a list of audit log entries pertaining to an item.
func (b *Builder) BuildGetAuditLogForItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	uri := b.BuildURL(
		ctx,
		nil,
		itemsBasePath,
		id(itemID),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}
