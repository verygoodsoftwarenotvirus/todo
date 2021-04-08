package items

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// ItemIDURIParamKey is a standard string that we'll use to refer to item IDs with.
	ItemIDURIParamKey = "itemID"
)

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

// CreateHandler is our item creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.ItemCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID).WithValue(keys.AccountIDKey, reqCtx.ActiveAccountID)
	input.BelongsToAccount = reqCtx.ActiveAccountID

	// create item in database.
	item, err := s.itemDataManager.CreateItem(ctx, input, reqCtx.Requester.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachItemIDToSpan(span, item.ID)
	logger = logger.WithValue(keys.ItemIDKey, item.ID)

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, item.ID, item); searchIndexErr != nil {
		observability.AcknowledgeError(err, logger, span, "adding item to search index")
	}

	s.itemCounter.Increment(ctx)
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, item, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an item.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID).WithValue(keys.AccountIDKey, reqCtx.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, reqCtx.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

// ExistenceHandler returns a HEAD handler that returns 200 if an item exists, 404 otherwise.
func (s *service) ExistenceHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		s.logger.Error(err, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID).WithValue(keys.AccountIDKey, reqCtx.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	exists, err := s.itemDataManager.ItemExists(ctx, itemID, reqCtx.ActiveAccountID)
	if !errors.Is(err, sql.ErrNoRows) {
		observability.AcknowledgeError(err, logger, span, "checking item existence")
	}

	if !exists || errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
	}
}

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	filter := types.ExtractQueryFilter(req)
	logger := s.logger.WithRequest(req).
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterPageKey, filter.Page).
		WithValue(keys.FilterSortByKey, string(filter.SortBy))

	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := reqCtx.Requester.ServiceAdminPermissions.IsServiceAdmin() && adminQueryPresent

	var items *types.ItemList

	if reqCtx.Requester.ServiceAdminPermissions.IsServiceAdmin() && isAdminRequest {
		items, err = s.itemDataManager.GetItemsForAdmin(ctx, filter)
	} else {
		items, err = s.itemDataManager.GetItems(ctx, reqCtx.ActiveAccountID, filter)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = &types.ItemList{Items: []*types.Item{}}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, items)
}

// SearchHandler is our search route.
func (s *service) SearchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	query := req.URL.Query().Get(types.SearchQueryKey)
	filter := types.ExtractQueryFilter(req)
	logger := s.logger.WithRequest(req).
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterPageKey, filter.Page).
		WithValue(keys.FilterSortByKey, string(filter.SortBy)).
		WithValue(keys.SearchQueryKey, query)

	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		s.logger.Error(err, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := reqCtx.Requester.ServiceAdminPermissions.IsServiceAdmin() && adminQueryPresent

	var (
		relevantIDs []uint64
		items       []*types.Item
	)

	if isAdminRequest {
		relevantIDs, err = s.search.SearchForAdmin(ctx, query)
	} else {
		relevantIDs, err = s.search.Search(ctx, query, reqCtx.ActiveAccountID)
	}

	if err != nil {
		observability.AcknowledgeError(err, logger, span, "executing item search query")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// fetch items from database.
	if isAdminRequest {
		items, err = s.itemDataManager.GetItemsWithIDsForAdmin(ctx, filter.Limit, relevantIDs)
	} else {
		items, err = s.itemDataManager.GetItemsWithIDs(ctx, reqCtx.ActiveAccountID, filter.Limit, relevantIDs)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = []*types.Item{}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "searching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, items)
}

// UpdateHandler returns a handler that updates an item.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check for parsed input attached to request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.ItemUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID).WithValue(keys.AccountIDKey, reqCtx.ActiveAccountID)
	input.BelongsToAccount = reqCtx.ActiveAccountID

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, reqCtx.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving item for update")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := x.Update(input)
	tracing.AttachChangeSummarySpan(span, "item", changeReport)

	// update item in database.
	if err = s.itemDataManager.UpdateItem(ctx, x, reqCtx.Requester.ID, changeReport); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
		observability.AcknowledgeError(err, logger, span, "updating item in search index")
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an item.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID).WithValue(keys.AccountIDKey, reqCtx.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// archive the item in the database.
	err = s.itemDataManager.ArchiveItem(ctx, itemID, reqCtx.ActiveAccountID, reqCtx.Requester.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.itemCounter.Decrement(ctx)

	if indexDeleteErr := s.search.Delete(ctx, itemID); indexDeleteErr != nil {
		observability.AcknowledgeError(err, logger, span, "removing from search index")
	}

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	x, err := s.itemDataManager.GetAuditLogEntriesForItem(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving audit log entries for item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}
