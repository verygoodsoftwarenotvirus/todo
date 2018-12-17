package items

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/go-chi/chi"
)

const (
	URIParamKey = "itemID"
)

func (s *ItemsService) ItemContextMiddleware(next http.Handler) http.Handler {
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

func (s *ItemsService) Read(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, URIParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := s.db.GetItem(itemID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Errorf("error fetching item #%s from database: %v", itemIDParam, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (s *ItemsService) Count(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	itemCount, err := s.db.GetItemCount(qf)
	if err != nil {
		s.logger.Errorf("error fetching item count from database: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-type", "application/json")

	json.NewEncoder(res).Encode(struct {
		Count uint64 `json:"count"`
	}{itemCount})
}

func (s *ItemsService) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	items, err := s.db.GetItems(qf)
	if err != nil {
		s.logger.Errorln("error encountered fetching items: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(items)
}

func (s *ItemsService) Delete(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, URIParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	if err := s.db.DeleteItem(itemID); err != nil {
		s.logger.Errorf("error encountered deleting item %d: %v", itemID, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Update is our item update route
// note that Update is meant to happen after ItemContextMiddleware
func (s *ItemsService) Update(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		s.logger.Errorln("no input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	itemIDParam := chi.URLParam(req, URIParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := s.db.GetItem(itemID)
	if err != nil {
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

// Create is our item creation route
// note that Create is meant to happen after ItemContextMiddleware
func (s *ItemsService) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := s.db.CreateItem(input)
	if err != nil {
		s.logger.Errorf("error creating item: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}
