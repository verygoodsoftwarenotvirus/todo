package items

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
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

// ListHandler is our list route.
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("ListHandler invoked")

	// ensure query filter.
	filter := models.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := si.UserIsAdmin && adminQueryPresent

	var (
		items *models.ItemList
		err   error
	)

	if si.UserIsAdmin && isAdminRequest {
		items, err = s.itemDataManager.GetItemsForAdmin(ctx, filter)
	} else {
		items, err = s.itemDataManager.GetItems(ctx, si.UserID, filter)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = &models.ItemList{Items: []models.Item{}}
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, items)
}

// SearchHandler is our search route.
func (s *Service) SearchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "SearchHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("items search handler hit")

	// we only parse the filter here because it will contain the limit
	filter := models.ExtractQueryFilter(req)
	query := req.URL.Query().Get(models.SearchQueryKey)
	logger = logger.WithValue("search_query", query)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := si.UserIsAdmin && adminQueryPresent

	var (
		relevantIDs []uint64
		searchErr   error

		items []models.Item
		dbErr error
	)

	if isAdminRequest {
		relevantIDs, searchErr = s.search.SearchForAdmin(ctx, query)
	} else {
		relevantIDs, searchErr = s.search.Search(ctx, query, si.UserID)
	}

	if searchErr != nil {
		logger.Error(searchErr, "error encountered executing search query")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// fetch items from database.
	if isAdminRequest {
		items, dbErr = s.itemDataManager.GetItemsWithIDsForAdmin(ctx, filter.Limit, relevantIDs)
	} else {
		items, dbErr = s.itemDataManager.GetItemsWithIDs(ctx, si.UserID, filter.Limit, relevantIDs)
	}

	if errors.Is(dbErr, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		items = []models.Item{}
	} else if dbErr != nil {
		logger.Error(dbErr, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, items)
}

// CreateHandler is our item creation route.
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "CreateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*models.ItemCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)
		return
	}

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)
	input.BelongsToUser = si.UserID

	// create item in database.
	x, err := s.itemDataManager.CreateItem(ctx, input)
	if err != nil {
		logger.Error(err, "error creating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	tracing.AttachItemIDToSpan(span, x.ID)
	logger = logger.WithValue("item_id", x.ID)

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
		logger.Error(searchIndexErr, "adding item to search index")
	}

	s.itemCounter.Increment(ctx)
	s.auditLog.LogItemCreationEvent(ctx, x)
	s.encoderDecoder.EncodeResponseWithStatus(res, x, http.StatusCreated)
}

// ExistenceHandler returns a HEAD handler that returns 200 if an item exists, 404 otherwise.
func (s *Service) ExistenceHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ExistenceHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue("item_id", itemID)

	// fetch item from database.
	exists, err := s.itemDataManager.ItemExists(ctx, itemID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Error(err, "error checking item existence in database")
	}

	if !exists || errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
	}
}

// ReadHandler returns a GET handler that returns an item.
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue("item_id", itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching item from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}

// AuditEntryHandler returns a GET handler that returns an item.
func (s *Service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "AuditEntryHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue("item_id", itemID)

	x, err := s.auditLog.GetAuditLogEntriesForItem(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}

// UpdateHandler returns a handler that updates an item.
func (s *Service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "UpdateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check for parsed input attached to request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*models.ItemUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)
		return
	}

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)
	input.BelongsToUser = si.UserID

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue("item_id", itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// update the data structure.
	changeReport := x.Update(input)

	// update item in database.
	if err = s.itemDataManager.UpdateItem(ctx, x); err != nil {
		logger.Error(err, "error encountered updating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// notify relevant parties.
	if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
		logger.Error(searchIndexErr, "updating item in search index")
	}

	s.auditLog.LogItemUpdateEvent(ctx, si.UserID, x.ID, changeReport)

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}

// ArchiveHandler returns a handler that archives an item.
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ArchiveHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	logger = logger.WithValue("item_id", itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	// archive the item in the database.
	err := s.itemDataManager.ArchiveItem(ctx, itemID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// notify relevant parties.
	s.itemCounter.Decrement(ctx)
	s.auditLog.LogItemArchiveEvent(ctx, si.UserID, itemID)

	if indexDeleteErr := s.search.Delete(ctx, itemID); indexDeleteErr != nil {
		logger.Error(indexDeleteErr, "error removing item from search index")
	}

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}
