package items

import (
	"database/sql"
	"errors"
	"net/http"

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
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID).WithValue(keys.AccountIDKey, si.User.ActiveAccountID)
	input.BelongsToAccount = si.User.ActiveAccountID

	// create item in database.
	x, err := s.itemDataManager.CreateItem(ctx, input, si.User.ID)
	if err != nil {
		logger.Error(err, "error creating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachItemIDToSpan(span, x.ID)
	logger = logger.WithValue(keys.ItemIDKey, x.ID)

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
		logger.Error(searchIndexErr, "adding item to search index")
	}

	s.itemCounter.Increment(ctx)
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, x, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an item.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID).WithValue(keys.AccountIDKey, si.User.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, si.User.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching item from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ExistenceHandler returns a HEAD handler that returns 200 if an item exists, 404 otherwise.
func (s *service) ExistenceHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID).WithValue(keys.AccountIDKey, si.User.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	exists, err := s.itemDataManager.ItemExists(ctx, itemID, si.User.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Error(err, "error checking item existence in database")
	}

	if !exists || errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
	}
}

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("ListHandler invoked")

	// ensure query filter.
	filter := types.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := si.User.ServiceAdminPermissions.IsServiceAdmin() && adminQueryPresent

	var (
		items *types.ItemList
		err   error
	)

	if si.User.ServiceAdminPermissions.IsServiceAdmin() && isAdminRequest {
		items, err = s.itemDataManager.GetItemsForAdmin(ctx, filter)
	} else {
		items, err = s.itemDataManager.GetItems(ctx, si.User.ActiveAccountID, filter)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = &types.ItemList{Items: []*types.Item{}}
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, items)
}

// SearchHandler is our search route.
func (s *service) SearchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("items search handler hit")

	// we only parse the filter here because it will contain the limit
	filter := types.ExtractQueryFilter(req)
	query := req.URL.Query().Get(types.SearchQueryKey)
	logger = logger.WithValue(keys.SearchQueryKey, query)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := si.User.ServiceAdminPermissions.IsServiceAdmin() && adminQueryPresent

	var (
		relevantIDs []uint64
		searchErr   error

		items []*types.Item
		dbErr error
	)

	if isAdminRequest {
		relevantIDs, searchErr = s.search.SearchForAdmin(ctx, query)
	} else {
		relevantIDs, searchErr = s.search.Search(ctx, query, si.User.ActiveAccountID)
	}

	if searchErr != nil {
		logger.Error(searchErr, "error encountered executing search query")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// fetch items from database.
	if isAdminRequest {
		items, dbErr = s.itemDataManager.GetItemsWithIDsForAdmin(ctx, filter.Limit, relevantIDs)
	} else {
		items, dbErr = s.itemDataManager.GetItemsWithIDs(ctx, si.User.ActiveAccountID, filter.Limit, relevantIDs)
	}

	if errors.Is(dbErr, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = []*types.Item{}
	} else if dbErr != nil {
		logger.Error(dbErr, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, items)
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
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID).WithValue(keys.AccountIDKey, si.User.ActiveAccountID)
	input.BelongsToAccount = si.User.ActiveAccountID

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, si.User.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := x.Update(input)

	// update item in database.
	if err = s.itemDataManager.UpdateItem(ctx, x, changeReport, si.User.ID); err != nil {
		logger.Error(err, "error encountered updating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
		logger.Error(searchIndexErr, "updating item in search index")
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an item.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID).WithValue(keys.AccountIDKey, si.User.ActiveAccountID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// archive the item in the database.
	err := s.itemDataManager.ArchiveItem(ctx, itemID, si.User.ActiveAccountID, si.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.itemCounter.Decrement(ctx)

	if indexDeleteErr := s.search.Delete(ctx, itemID); indexDeleteErr != nil {
		logger.Error(indexDeleteErr, "error removing item from search index")
	}

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	x, err := s.itemDataManager.GetAuditLogEntriesForItem(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}
