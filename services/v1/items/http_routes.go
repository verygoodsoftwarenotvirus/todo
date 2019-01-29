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
		s.logger.Debugln("itemsService.ReadHandler called for item #", itemID)

		userID := s.userIDFetcher(req)
		i, err := s.db.GetItem(itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			s.logger.Errorf("error fetching item #%d from database: %v", itemID, err)
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
	itemCount, err := s.db.GetItemCount(qf)
	if err != nil {
		s.logger.Errorf("error fetching item count from database: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(&models.CountResponse{Count: itemCount})
}

// List is our list route
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ExtractQueryFilter(req)
	items, err := s.db.GetItems(qf)
	if err != nil {
		s.logger.Errorln("error encountered fetching items: ", err)
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
		s.logger.Debugln("ItemsService Deletion handler called")
		itemID := itemIDFetcher(req)
		err := s.db.DeleteItem(itemID)

		s.logger.Debugf("itemID: %d, err: %v", itemID, err)

		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			s.logger.Errorf("error encountered deleting item %d: %v", itemID, err)
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
		rctx := req.Context()
		input, ok := rctx.Value(MiddlewareCtxKey).(*models.ItemInput)
		if !ok {
			s.logger.Errorln("no input attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		userID := s.userIDFetcher(req)
		itemID := itemIDFetcher(req)
		i, err := s.db.GetItem(itemID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			s.logger.Errorf("error encountered getting item %d: %v", itemID, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		i.Update(input)
		if err := s.db.UpdateItem(i); err != nil {
			s.logger.Errorf("error encountered updating item %d: %v", itemID, err)
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
	rctx := req.Context()
	input, ok := rctx.Value(MiddlewareCtxKey).(*models.ItemInput)
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

	i, err := s.db.CreateItem(input)
	if err != nil {
		s.logger.Errorf("error creating item: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}
