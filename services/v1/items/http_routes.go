package items

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// URIParamKey is a standard string that we'll use to refer to item IDs with
	URIParamKey = "itemID"
)

// ItemInputMiddleware is a middleware for fetching, parsing, and attaching a parsed ItemInput struct from a request
func (s *Service) ItemInputMiddleware(next http.Handler) http.Handler {
	s.logger.Debugln("ItemInputMiddleware called")
	x := new(models.ItemInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// BuildReadHandler returns a GET handler that returns an item
func (s *Service) BuildReadHandler(itemIDFetcher func(*http.Request) uint64) http.HandlerFunc {
	if itemIDFetcher == nil {
		panic("itemIDFetcher provided to BuildRead cannot be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		itemID := itemIDFetcher(req)
		logger := s.logger.WithField("item_id", itemID)
		logger.Debugln("itemsService.ReadHandler called")

		userID := s.userIDFetcher(req)
		logger = logger.WithField("user_id", userID)
		ctx := req.Context()

		i, err := s.db.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			logger.Debugln("No rows found in database")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.WithError(err).Errorln("Error fetching item from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(i)
	}
}

// Count is our count route
func (s *Service) Count(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("ItemsService.Count called")
	qf := models.ExtractQueryFilter(req)

	logger := s.logger.WithField("filter", qf)
	ctx := req.Context()

	itemCount, err := s.db.GetItemCount(ctx, qf)
	if err != nil {
		logger.Errorf("error fetching item count from database: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(&models.CountResponse{Count: itemCount})
}

// List is our list route
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("ItemsService.List called")
	qf := models.ExtractQueryFilter(req)

	logger := s.logger.WithField("filter", qf)
	ctx := req.Context()

	items, err := s.db.GetItems(ctx, qf)
	if err != nil {
		logger.Errorln("error encountered fetching items: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(items)
}

// BuildDeleteHandler returns a handler that deletes an item
func (s *Service) BuildDeleteHandler(itemIDFetcher func(*http.Request) uint64) http.HandlerFunc {
	if itemIDFetcher == nil {
		panic("itemIDFetcher provided to BuildRead cannot be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		itemID := itemIDFetcher(req)
		logger := s.logger.WithField("item_id", itemID)
		logger.Debugln("ItemsService Deletion handler called")

		err := s.db.DeleteItem(ctx, itemID)
		if err == sql.ErrNoRows {
			logger.Debugln("no rows found for item")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.WithError(err).Errorln("error encountered deleting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// BuildUpdateHandler returns a handler that updates an item
func (s *Service) BuildUpdateHandler(itemIDFetcher func(*http.Request) uint64) http.HandlerFunc {
	if itemIDFetcher == nil {
		panic("itemIDFetcher provided to BuildRead cannot be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		input, ok := ctx.Value(MiddlewareCtxKey).(*models.ItemInput)
		if !ok {
			s.logger.Errorln("no input attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		userID := s.userIDFetcher(req)
		itemID := itemIDFetcher(req)

		logger := s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"item_id": itemID,
			"input":   input,
		})

		i, err := s.db.GetItem(ctx, itemID, userID)
		if err == sql.ErrNoRows {
			logger.Debugln("no rows found for item")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.WithError(err).Errorln("error encountered getting item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		i.Update(input)
		if err := s.db.UpdateItem(ctx, i); err != nil {
			logger.WithError(err).Errorln("error encountered updating item")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(i)
	}
}

// Create is our item creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("ItemsService.Create called")
	ctx := req.Context()
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"input == nil":           input == nil,
		"s.userIDFetcher == nil": s.userIDFetcher == nil,
		"req == nil":             req == nil,
	}).Debugln("ItemsService.Create called")

	input.BelongsTo = s.userIDFetcher(req)

	i, err := s.db.CreateItem(ctx, input)
	if err != nil {
		s.logger.Errorf("error creating item: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}
