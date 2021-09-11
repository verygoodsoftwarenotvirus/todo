package items

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// ItemIDURIParamKey is a standard string that we'll use to refer to item IDs with.
	ItemIDURIParamKey = "itemID"
)

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "1", "t", "true":
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
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// check session context data for parsed input struct.
	input := new(types.ItemDatabaseCreationInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	input.ID = ksuid.New().String()
	input.BelongsToAccount = sessionCtxData.ActiveAccountID

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	pendingWrite := &workers.PendingWriteMessage{
		MessageType:          "item",
		Item:                 input,
		AttributableToUserID: sessionCtxData.Requester.UserID,
	}
	if err = s.pendingWritesProducer.Publish(ctx, pendingWrite); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing item pending write")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	pwr := types.PendingWriteResponse{ID: input.ID}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, pwr, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an item.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	x, err := s.itemDataManager.GetItem(ctx, itemID, sessionCtxData.ActiveAccountID)
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
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		s.logger.Error(err, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// check the database.
	exists, err := s.itemDataManager.ItemExists(ctx, itemID, sessionCtxData.ActiveAccountID)
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

	tracing.AttachRequestToSpan(span, req)
	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	items, err := s.itemDataManager.GetItems(ctx, sessionCtxData.ActiveAccountID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist, return an empty list.
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

	tracing.AttachRequestToSpan(span, req)
	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		s.logger.Error(err, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	relevantIDs, err := s.search.Search(ctx, query, sessionCtxData.ActiveAccountID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "executing item search query")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// fetch items from database.
	items, err := s.itemDataManager.GetItemsWithIDs(ctx, sessionCtxData.ActiveAccountID, filter.Limit, relevantIDs)
	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist, return an empty list.
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
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// check for parsed input attached to session context data.
	input := new(types.ItemUpdateInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		logger.Error(err, "error encountered decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.Error(err, "provided input was invalid")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}
	input.BelongsToAccount = sessionCtxData.ActiveAccountID

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// fetch item from database.
	item, err := s.itemDataManager.GetItem(ctx, itemID, sessionCtxData.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving item for update")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the item.
	changeReport := item.Update(input)
	tracing.AttachChangeSummarySpan(span, "item", changeReport)

	// update item in database.
	if err = s.itemDataManager.UpdateItem(ctx, item, sessionCtxData.Requester.UserID, changeReport); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify interested parties.
	if searchIndexErr := s.search.Index(ctx, item.ID, item); searchIndexErr != nil {
		observability.AcknowledgeError(err, logger, span, "updating item in search index")
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, item)
}

// ArchiveHandler returns a handler that archives an item.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// determine item ID.
	itemID := s.itemIDFetcher(req)
	tracing.AttachItemIDToSpan(span, itemID)
	logger = logger.WithValue(keys.ItemIDKey, itemID)

	// archive the item in the database.
	err = s.itemDataManager.ArchiveItem(ctx, itemID, sessionCtxData.ActiveAccountID, sessionCtxData.Requester.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving item")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify interested parties.
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
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

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
