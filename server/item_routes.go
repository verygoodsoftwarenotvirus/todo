package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/go-chi/chi"
)

func (s *Server) getItem(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, "itemID")
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := s.db.GetItem(uint(itemID))
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		s.Logger.Errorf("error fetching item #%s from database: %v", itemIDParam, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (s *Server) getItems(res http.ResponseWriter, req *http.Request) {
	items, err := s.db.GetItems(nil)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(items)
}

func (s *Server) deleteItem(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, "itemID")
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	if err := s.db.DeleteItem(uint(itemID)); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateItem(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(models.ItemInputCtxKey).(*models.ItemInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	itemIDParam := chi.URLParam(req, "itemID")
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := s.db.GetItem(uint(itemID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	i.Update(input)
	if err := s.db.UpdateItem(i); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (s *Server) createItem(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(models.ItemInputCtxKey).(*models.ItemInput)
	if !ok {
		s.Logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := s.db.CreateItem(input)
	if err != nil {
		s.Logger.Errorf("error creating item: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}
