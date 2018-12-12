package server

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauthclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/context"
)

func chiUsernameFetcher(req *http.Request) string {
	// PONDER: if the only time we use users.URIParamKey is externally to the users package
	// does it really need to belong there?
	return chi.URLParam(req, users.URIParamKey)
}

func (s *Server) setupRoutes() {
	router := chi.NewRouter()
	router.Use(
		context.ClearHandler,
		middleware.RequestID,
		middleware.DefaultLogger,
		middleware.Timeout(maxTimeout),
	)
	router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) { res.WriteHeader(http.StatusOK) })

	if s.DebugMode {
		router.Get("/_debug_/stats", s.stats)
	}

	router.Route("/users", func(userRouter chi.Router) {
		sr := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", users.URIParamKey)
		userRouter.With(s.usersService.UserLoginInputContextMiddleware).Post("/login", s.Login)
		userRouter.Post("/logout", s.Logout)

		userRouter.
			With(
				s.UserAuthenticationMiddleware,
				s.usersService.TOTPSecretRefreshInputContextMiddleware).
			Post("/totp_secret/new", s.usersService.NewTOTPSecret(chiUsernameFetcher))
		userRouter.
			With(
				s.UserAuthenticationMiddleware,
				s.usersService.PasswordUpdateInputContextMiddleware).
			Post("/password/new", s.usersService.UpdatePassword(chiUsernameFetcher))

		userRouter.Get("/", s.usersService.List)     // List
		userRouter.Get(sr, s.usersService.Read)      // Read
		userRouter.Delete(sr, s.usersService.Delete) // Delete
		// userRouter.With(s.usersService.UserInputContextMiddleware).Put(sr, s.usersService.Update)   // Update
		userRouter.With(s.usersService.UserInputContextMiddleware).Post("/", s.usersService.Create) // Create
	})

	router.Route("/oauth2", func(oauthRouter chi.Router) {
		// pk := fmt.Sprintf(`{%s:[a-zA-Z0-9\_\-]+}`, oauth2ClientIDURIParamKey)
		oauthRouter.
			With(s.UserAuthenticationMiddleware).
			Post("/clients", func(res http.ResponseWriter, req *http.Request) {

			})
	})

	router.
		With(
			s.UserAuthenticationMiddleware,
			s.Oauth2ClientInfoMiddleware,
		).
		Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2Handler.HandleAuthorizeRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
		})

	router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
		s.oauth2Handler.HandleTokenRequest(res, req)
	})

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {
			v1Router.With(s.UserAuthenticationMiddleware).Post("/fart", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusTeapot)
			})

			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
				itemsRouter.Get("/", s.itemsService.List)                                               // List
				itemsRouter.Get("/count", s.itemsService.Count)                                         // List
				itemsRouter.Get(sr, s.itemsService.Read)                                                // Read
				itemsRouter.Delete(sr, s.itemsService.Delete)                                           // Delete
				itemsRouter.With(s.itemsService.ItemContextMiddleware).Put(sr, s.itemsService.Update)   // Update
				itemsRouter.With(s.itemsService.ItemContextMiddleware).Post("/", s.itemsService.Create) // Create
			})

			v1Router.
				With(s.UserAuthenticationMiddleware).
				Route("/clients", func(oauth2ClientRouter chi.Router) {
					sr := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", oauthclients.URIParamKey)
					oauth2ClientRouter.Get("/", s.oauthclientsService.List)                                                                   // List
					oauth2ClientRouter.Get(sr, s.oauthclientsService.Read)                                                                    // Read
					oauth2ClientRouter.Delete(sr, s.oauthclientsService.Delete)                                                               // Delete
					oauth2ClientRouter.With(s.oauthclientsService.Oauth2ClientInputContextMiddleware).Put(sr, s.oauthclientsService.Update)   // Update
					oauth2ClientRouter.With(s.oauthclientsService.Oauth2ClientInputContextMiddleware).Post("/", s.oauthclientsService.Create) // Create
				})

		})
	})

	s.server.Handler = router
}
