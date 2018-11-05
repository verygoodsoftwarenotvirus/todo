package todo

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

const (
	maxTimeout = 120 * time.Second
)

type Server struct {
	certFile      string
	keyFile       string
	db            database.Database
	server        *http.Server
	Authenticator auth.Enticator
	DebugMode     bool
	Router        *chi.Mux
	Logger        *logrus.Logger
}

func buildServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  maxTimeout,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}
}

func setupRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.DefaultLogger)
	router.Use(middleware.Timeout(maxTimeout))

	return router
}

func New(certFile, keyFile string) *Server {
	return &Server{
		certFile: certFile,
		keyFile:  keyFile,
		server:   buildServer(),
		Logger:   logrus.New(),
		Router:   setupRouter(),
	}
}

func NewDebug(certFile, keyFile string) *Server {
	c := New(certFile, keyFile)
	c.DebugMode = true
	c.Logger.SetLevel(logrus.DebugLevel)
	return c
}

func (c *Server) Serve() {
	c.Logger.Debugf("About to listen on 443. Go to https://localhost/")
	c.server.Handler = c.Router
	c.Logger.Fatal(c.server.ListenAndServeTLS(c.certFile, c.keyFile))
}

func (s *Server) setupRoutes() *chi.Mux {
	router := setupRouter()

	router.Get("/item/{itemID:[0-9]+}", s.getItem)       // Read
	router.Get("/items", s.getItems)                     // List
	router.Patch("/item/{itemID:[0-9]+}", s.updateItem)  // Update
	router.Post("/items", s.createItem)                  // Create
	router.Delete("/item/{itemID:[0-9]+}", s.deleteItem) // Delete

	return router
}

func (s *Server) getItem(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, "itemID")
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	i, err := s.db.GetItem(uint(itemID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(res).Encode(i)
}

func (s *Server) getItems(res http.ResponseWriter, req *http.Request) {
	items, err := s.db.GetItems(nil)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(res).Encode(items)
}

func (s *Server) deleteItem(res http.ResponseWriter, req *http.Request) {
	itemIDParam := chi.URLParam(req, "itemID")
	itemID, _ := strconv.ParseUint(itemIDParam, 10, 64)

	if err := s.db.DeleteItem(uint(itemID)); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
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
	}

	i.Update(input)
	// TODO: merge these two values somehow

	json.NewEncoder(res).Encode(i)
}

func (s *Server) createItem(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(models.ItemInputCtxKey).(*models.ItemInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := s.db.CreateItem(input)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(res).Encode(i)
}
