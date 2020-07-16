package items

import (
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// URIParamKey is a standard string that we'll use to refer to item IDs with.
	URIParamKey = "itemID"
)

// ListHandler is our list route.
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// ensure query filter.
		filter := models.ExtractQueryFilter(req)

		// determine user ID.
		userID := s.userIDFetcher(req)
		tracing.AttachUserIDToSpan(span, userID)
		logger = logger.WithValue("user_id", userID)

		// fetch items from database.
		items, err := s.itemDataManager.GetItems(ctx, userID, filter)
		if err == sql.ErrNoRows {
			// in the event no rows exist return an empty list.
			items = &models.ItemList{
				Items: []models.Item{},
			}
		} else if err != nil {
			logger.Error(err, "error encountered fetching items")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode our response and peace.
		if err = s.encoderDecoder.EncodeResponse(res, items); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// SearchHandler is our list route.
func (s *Service) SearchHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "SearchHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// we only parse the filter here because it will contain the limit
		filter := models.ExtractQueryFilter(req)
		query := req.URL.Query().Get(models.SearchQueryKey)
		logger = logger.WithValue("search_query", query)

		// determine user ID.
		userID := s.userIDFetcher(req)
		tracing.AttachUserIDToSpan(span, userID)
		logger = logger.WithValue("user_id", userID)

		relevantIDs, searchErr := s.search.Search(ctx, query, userID)
		if searchErr != nil {
			logger.Error(searchErr, "error encountered executing search query")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// fetch items from database.
		items, err := s.itemDataManager.GetItemsWithIDs(ctx, userID, filter.Limit, relevantIDs)
		if err == sql.ErrNoRows {
			// in the event no rows exist return an empty list.
			items = []models.Item{}
		} else if err != nil {
			logger.Error(err, "error encountered fetching items")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode our response and peace.
		if err = s.encoderDecoder.EncodeResponse(res, items); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// CreateHandler is our item creation route.
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check request context for parsed input struct.
		input, ok := ctx.Value(CreateMiddlewareCtxKey).(*models.ItemCreationInput)
		if !ok {
			logger.Info("valid input not attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// determine user ID.
		userID := s.userIDFetcher(req)
		logger = logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)
		input.BelongsToUser = userID

		// create item in database.
		x, err := s.itemDataManager.CreateItem(ctx, input)
		if err != nil {
			logger.Error(err, "error creating item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracing.AttachItemIDToSpan(span, x.ID)
		logger = logger.WithValue("item_id", x.ID)

		// notify relevant parties.
		s.itemCounter.Increment(ctx)
		s.reporter.Report(newsman.Event{
			Data:      x,
			Topics:    []string{topicName},
			EventType: string(models.Create),
		})
		if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
			s.logger.Error(searchIndexErr, "adding item to search index")
		}

		// encode our response and peace.
		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ExistenceHandler returns a HEAD handler that returns 200 if an item exists, 404 otherwise.
func (s *Service) ExistenceHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ExistenceHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// determine user ID.
		userID := s.userIDFetcher(req)
		tracing.AttachUserIDToSpan(span, userID)
		logger = logger.WithValue("user_id", userID)

		// determine item ID.
		itemID := s.itemIDFetcher(req)
		tracing.AttachItemIDToSpan(span, itemID)
		logger = logger.WithValue("item_id", itemID)

		// fetch item from database.
		exists, err := s.itemDataManager.ItemExists(ctx, itemID, userID)
		if err != nil && err != sql.ErrNoRows {
			logger.Error(err, "error checking item existence in database")
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if exists {
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusNotFound)
		}
	}
}

// ReadHandler returns a GET handler that returns an item.
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// determine user ID.
		userID := s.userIDFetcher(req)
		tracing.AttachUserIDToSpan(span, userID)
		logger = logger.WithValue("user_id", userID)

		// determine item ID.
		itemID := s.itemIDFetcher(req)
		tracing.AttachItemIDToSpan(span, itemID)
		logger = logger.WithValue("item_id", itemID)

		// fetch item from database.
		x, err := s.itemDataManager.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error fetching item from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode our response and peace.
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// UpdateHandler returns a handler that updates an item.
func (s *Service) UpdateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "UpdateHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check for parsed input attached to request context.
		input, ok := ctx.Value(UpdateMiddlewareCtxKey).(*models.ItemUpdateInput)
		if !ok {
			logger.Info("no input attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// determine user ID.
		userID := s.userIDFetcher(req)
		logger = logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)
		input.BelongsToUser = userID

		// determine item ID.
		itemID := s.itemIDFetcher(req)
		logger = logger.WithValue("item_id", itemID)
		tracing.AttachItemIDToSpan(span, itemID)

		// fetch item from database.
		x, err := s.itemDataManager.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered getting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update the data structure.
		x.Update(input)

		// update item in database.
		if err = s.itemDataManager.UpdateItem(ctx, x); err != nil {
			logger.Error(err, "error encountered updating item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if searchIndexErr := s.search.Index(ctx, x.ID, x); searchIndexErr != nil {
			s.logger.Error(searchIndexErr, "adding item to search index")
		}

		// notify relevant parties.
		s.reporter.Report(newsman.Event{
			Data:      x,
			Topics:    []string{topicName},
			EventType: string(models.Update),
		})

		// encode our response and peace.
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ArchiveHandler returns a handler that archives an item.
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		ctx, span := tracing.StartSpan(req.Context(), "ArchiveHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// determine user ID.
		userID := s.userIDFetcher(req)
		logger = logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)

		// determine item ID.
		itemID := s.itemIDFetcher(req)
		logger = logger.WithValue("item_id", itemID)
		tracing.AttachItemIDToSpan(span, itemID)

		// archive the item in the database.
		err = s.itemDataManager.ArchiveItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered deleting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify relevant parties.
		s.itemCounter.Decrement(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Archive),
			Data:      &models.Item{ID: itemID},
			Topics:    []string{topicName},
		})
		if indexDeleteErr := s.search.Delete(ctx, itemID); indexDeleteErr != nil {
			s.logger.Error(indexDeleteErr, "error removing item from search index")
		}

		// encode our response and peace.
		res.WriteHeader(http.StatusNoContent)
	}
}
