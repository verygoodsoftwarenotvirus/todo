package server

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/context"
)

// chiUsernameFetcher (now named uf) fetches a username from a request routed by chi.
func uf(req *http.Request) string {
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

		userRouter.With(s.UserAuthenticationMiddleware, s.usersService.TOTPSecretRefreshInputContextMiddleware).
			Post("/totp_secret/new", s.usersService.NewTOTPSecret(uf))

		userRouter.With(s.UserAuthenticationMiddleware, s.usersService.PasswordUpdateInputContextMiddleware).
			Post("/password/new", s.usersService.UpdatePassword(uf))

		userRouter.Get("/", s.usersService.List)     // List
		userRouter.Get(sr, s.usersService.Read)      // Read
		userRouter.Delete(sr, s.usersService.Delete) // Delete
		// userRouter.With(s.usersService.UserInputContextMiddleware).Put(sr, s.usersService.Update)   // Update
		userRouter.With(s.usersService.UserInputContextMiddleware).Post("/", s.usersService.Create) // Create
	})

	router.Route("/oauth2", func(oauth2Router chi.Router) {
		// pk := fmt.Sprintf(`{%s:[a-zA-Z0-9\_\-]+}`, oauth2ClientIDURIParamKey)
		oauth2Router.
			With(s.UserAuthenticationMiddleware).
			Post("/clients", func(res http.ResponseWriter, req *http.Request) {

			})

		oauth2Router.
			With(
				s.UserAuthenticationMiddleware,
				s.Oauth2ClientInfoMiddleware,
			).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
				err := s.oauth2Handler.HandleAuthorizeRequest(res, req)
				if err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
				}
			})

		oauth2Router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
			s.oauth2Handler.HandleTokenRequest(res, req)
		})

	})

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {

			v1Router.With(
				//s.UserAuthenticationMiddleware,
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
						token, err := s.oauth2Handler.ValidationBearerToken(req)
						if err != nil {
							http.Error(res, err.Error(), http.StatusUnauthorized)
							return
						}
						_ = token
					})
				},
			).Post("/fart", func(res http.ResponseWriter, req *http.Request) {
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

			v1Router.Route("/oauth2", func(oauth2Router chi.Router) {
				oauth2Router.Route("/clients", func(clientRouter chi.Router) {
					sr := fmt.Sprintf("/{%s}", oauthclients.URIParamKey)
					clientRouter.Get("/", s.oauth2ClientsService.List)     // List
					clientRouter.Get(sr, s.oauth2ClientsService.Read)      // Read
					clientRouter.Delete(sr, s.oauth2ClientsService.Delete) // Delete
					clientRouter.
						With(s.oauth2ClientsService.Oauth2ClientInputContextMiddleware).
						Put(sr, s.oauth2ClientsService.Update) // Update
					clientRouter.
						With(s.oauth2ClientsService.Oauth2ClientInputContextMiddleware).
						Post("/", s.oauth2ClientsService.Create) // Create
				})
			})
		})
	})

	s.server.Handler = router
}
