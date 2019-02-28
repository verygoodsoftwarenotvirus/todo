package items

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (
	// URIParamKey is a standard string that we'll use to refer to item IDs with
	URIParamKey = "itemID"
)

// ItemInputMiddleware is a middleware for fetching, parsing, and attaching a parsed ItemInput struct from a request
func (s *Service) ItemInputMiddleware(next http.Handler) http.Handler {
	x := new(models.ItemInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
		ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
		defer serverSpan.Finish()

		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s.logger.WithValue("itemInput", x).Debug("ItemInputMiddleware called")
		ctx = context.WithValue(ctx, MiddlewareCtxKey, x)

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// Read returns a GET handler that returns an item
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	defer serverSpan.Finish()

	userID := s.userIDFetcher(req)
	itemID := s.itemIDFetcher(req)

	logger := s.logger.WithValues(map[string]interface{}{
		"user_id": userID,
		"item_id": itemID,
	})
	logger.Debug("itemsService.ReadHandler called")

	i, err := s.db.GetItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("No rows found in database")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "Error fetching item from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, i); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Count is our count route
func (s *Service) Count(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	defer serverSpan.Finish()

	s.logger.Debug("ItemsService.Count called")
	qf := models.ExtractQueryFilter(req)

	logger := s.logger.WithValue("filter", qf)
	userID := s.userIDFetcher(req)

	itemCount, err := s.db.GetItemCount(ctx, qf, userID)
	logger = logger.WithValue("item_count", itemCount)
	if err != nil {
		logger.Error(err, "error fetching item count from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := &models.CountResponse{Count: itemCount}

	if err = s.encoder.EncodeResponse(res, c); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// List is our list route
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	defer serverSpan.Finish()

	s.logger.Debug("ItemsService.List called")
	qf := models.ExtractQueryFilter(req)

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"filter":  qf,
		"user_id": userID,
	})

	items, err := s.db.GetItems(ctx, qf, userID)
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

// Delete returns a handler that deletes an item
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	userID := s.userIDFetcher(req)
	itemID := s.itemIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	})
	logger.Debug("ItemsService Deletion handler called")

	err := s.db.DeleteItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for item")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

// Update returns a handler that updates an item
func (s *Service) Update(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

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

	i, err := s.db.GetItem(ctx, itemID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for item")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	i.Update(input)
	if err = s.db.UpdateItem(ctx, i); err != nil {
		logger.Error(err, "error encountered updating item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, i); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Create is our item creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	s.logger.Debug("ItemsService.Create called")
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		s.logger.Error(nil, "valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	input.BelongsTo = s.userIDFetcher(req)

	s.logger.WithValue("input", input).Debug("ItemsService.Create called")
	i, err := s.db.CreateItem(ctx, input)
	if err != nil {
		s.logger.Error(err, "error creating item")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, i); err != nil {
		s.logger.Error(err, "encoding response")
	}
}
