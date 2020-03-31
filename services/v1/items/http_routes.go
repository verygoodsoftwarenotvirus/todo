package items

import (
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"go.opencensus.io/trace"
)

const (
	// URIParamKey is a standard string that we'll use to refer to item IDs with
	URIParamKey = "itemID"
)

// ListHandler is our list route
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		// ensure query filter
		filter := models.ExtractQueryFilter(req)

		// determine user ID
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)

		// fetch items from database
		items, err := s.itemDatabase.GetItems(ctx, userID, filter)
		if err == sql.ErrNoRows {
			// in the event no rows exist return an empty list
			items = &models.ItemList{
				Items: []models.Item{},
			}
		} else if err != nil {
			logger.Error(err, "error encountered fetching items")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode our response and peace
		if err = s.encoderDecoder.EncodeResponse(res, items); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// CreateHandler is our item creation route
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		// determine user ID
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)

		// check request context for parsed input struct
		input, ok := ctx.Value(CreateMiddlewareCtxKey).(*models.ItemCreationInput)
		if !ok {
			logger.Info("valid input not attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		input.BelongsToUser = userID
		logger = logger.WithValue("input", input)

		// create item in database
		x, err := s.itemDatabase.CreateItem(ctx, input)
		if err != nil {
			logger.Error(err, "error creating item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify relevant parties
		s.itemCounter.Increment(ctx)
		tracing.AttachItemIDToSpan(span, x.ID)
		s.reporter.Report(newsman.Event{
			Data:      x,
			Topics:    []string{topicName},
			EventType: string(models.Create),
		})

		// encode our response and peace
		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// ExistenceHandler returns a HEAD handler that returns 200 if an item exists, 404 otherwise
func (s *Service) ExistenceHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ExistenceHandler")
		defer span.End()

		// determine relevant information
		userID := s.userIDFetcher(req)
		itemID := s.itemIDFetcher(req)
		logger := s.logger.WithValues(map[string]interface{}{
			"user_id": userID,
			"item_id": itemID,
		})
		tracing.AttachItemIDToSpan(span, itemID)
		tracing.AttachUserIDToSpan(span, userID)

		// fetch item from database
		exists, err := s.itemDatabase.ItemExists(ctx, itemID, userID)
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

// ReadHandler returns a GET handler that returns an item
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		// determine relevant information
		userID := s.userIDFetcher(req)
		itemID := s.itemIDFetcher(req)
		logger := s.logger.WithValues(map[string]interface{}{
			"user_id": userID,
			"item_id": itemID,
		})
		tracing.AttachItemIDToSpan(span, itemID)
		tracing.AttachUserIDToSpan(span, userID)

		// fetch item from database
		x, err := s.itemDatabase.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error fetching item from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode our response and peace
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// UpdateHandler returns a handler that updates an item
func (s *Service) UpdateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "UpdateHandler")
		defer span.End()

		// check for parsed input attached to request context
		input, ok := ctx.Value(UpdateMiddlewareCtxKey).(*models.ItemUpdateInput)
		if !ok {
			s.logger.Info("no input attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// determine relevant information
		userID := s.userIDFetcher(req)
		itemID := s.itemIDFetcher(req)
		logger := s.logger.WithValues(map[string]interface{}{
			"user_id": userID,
			"item_id": itemID,
		})
		tracing.AttachItemIDToSpan(span, itemID)
		tracing.AttachUserIDToSpan(span, userID)

		// fetch item from database
		x, err := s.itemDatabase.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered getting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update the data structure
		x.Update(input)

		// update item in database
		if err = s.itemDatabase.UpdateItem(ctx, x); err != nil {
			logger.Error(err, "error encountered updating item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify relevant parties
		s.reporter.Report(newsman.Event{
			Data:      x,
			Topics:    []string{topicName},
			EventType: string(models.Update),
		})

		// encode our response and peace
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// ArchiveHandler returns a handler that archives an item
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ArchiveHandler")
		defer span.End()

		// determine relevant information
		userID := s.userIDFetcher(req)
		itemID := s.itemIDFetcher(req)
		logger := s.logger.WithValues(map[string]interface{}{
			"item_id": itemID,
			"user_id": userID,
		})
		tracing.AttachItemIDToSpan(span, itemID)
		tracing.AttachUserIDToSpan(span, userID)

		// archive the item in the database
		err := s.itemDatabase.ArchiveItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered deleting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify relevant parties
		s.itemCounter.Decrement(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Archive),
			Data:      &models.Item{ID: itemID},
			Topics:    []string{topicName},
		})

		// encode our response and peace
		res.WriteHeader(http.StatusNoContent)
	}
}
