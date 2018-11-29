package items

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/go-chi/chi"
)

const (
	urlParamKey = "itemID"
)

func (is *ItemsService) ItemContextMiddleware(next http.Handler) http.Handler {
	x := new(models.ItemInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			is.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (is *ItemsService) Read(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, urlParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := is.db.GetItem(uint(itemID))
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		is.logger.Errorf("error fetching item #%s from database: %v", itemIDParam, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (is *ItemsService) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	items, err := is.db.GetItems(qf)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(items)
}

func (is *ItemsService) Delete(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, urlParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	if err := is.db.DeleteItem(uint(itemID)); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (is *ItemsService) Update(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	itemIDParam := chi.URLParam(req, urlParamKey)
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := is.db.GetItem(uint(itemID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	i.Update(input)
	if err := is.db.UpdateItem(i); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (is *ItemsService) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.ItemInput)
	if !ok {
		is.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := is.db.CreateItem(input)
	if err != nil {
		is.logger.Errorf("error creating item: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}
