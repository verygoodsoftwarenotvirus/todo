package items

import (
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"go.opencensus.io/trace"
)

const (
	// URIParamKey is a standard string that we'll use to refer to item IDs with
	URIParamKey = "itemID"
)

// List is our list route
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "list_route")
	defer span.End()

	s.logger.Debug("ItemsService.List called")
	qf := models.ExtractQueryFilter(req)

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"filter":  qf,
		"user_id": userID,
	})

	items, err := s.itemDatabase.GetItems(ctx, qf, userID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, items); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Create is our item creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "create_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)

	logger.Debug("create route called")
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.ItemInput)
	logger = logger.WithValue("input", input)
	if !ok {
		s.logger.Error(nil, "valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	input.BelongsTo = userID

	x, err := s.itemDatabase.CreateItem(ctx, input)
	if err != nil {
		s.logger.Error(err, "error creating item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.itemCounter.Increment(ctx)

	s.newsman.Report(newsman.Event{
		EventType: string(events.Create),
		Data:      x,
		Topics:    []string{topicName},
	})

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Read returns a GET handler that returns an item
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "read_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	itemID := s.itemIDFetcher(req)

	logger := s.logger.WithValues(map[string]interface{}{
		"user_id": userID,
		"item_id": itemID,
	})
	logger.Debug("itemsService.ReadHandler called")

	i, err := s.itemDatabase.GetItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("No rows found in itemDatabase")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "Error fetching item from itemDatabase")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, i); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Update returns a handler that updates an item
func (s *Service) Update(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "update_route")
	defer span.End()

	input, ok := ctx.Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		s.logger.Error(nil, "no input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := s.userIDFetcher(req)
	itemID := s.itemIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"user_id": userID,
		"item_id": itemID,
		"input":   input,
	})

	x, err := s.itemDatabase.GetItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for item")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	x.Update(input)
	if err = s.itemDatabase.UpdateItem(ctx, x); err != nil {
		logger.Error(err, "error encountered updating item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.newsman.Report(newsman.Event{
		EventType: string(events.Update),
		Data:      x,
		Topics:    []string{topicName},
	})

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Delete returns a handler that deletes an item
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "delete_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	itemID := s.itemIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	})
	logger.Debug("ItemsService deletion handler called")

	err := s.itemDatabase.DeleteItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for item")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.itemCounter.Decrement(ctx)

	s.newsman.Report(newsman.Event{
		EventType: string(events.Delete),
		Data:      &models.Item{ID: itemID},
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusNoContent)
}
