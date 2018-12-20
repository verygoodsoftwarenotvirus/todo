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
		usernamePattern := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", users.URIParamKey)
		userRouter.With(s.usersService.UserLoginInputContextMiddleware).Post("/login", s.Login)
		userRouter.With(s.UserCookieAuthenticationMiddleware).Post("/logout", s.Logout)

		userRouter.With(
			s.UserCookieAuthenticationMiddleware,
			s.usersService.TOTPSecretRefreshInputContextMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecret(chiUsernameFetcher))

		userRouter.With(
			s.UserCookieAuthenticationMiddleware,
			s.usersService.PasswordUpdateInputContextMiddleware,
		).Post("/password/new", s.usersService.UpdatePassword(chiUsernameFetcher))

		userRouter.Get("/", s.usersService.List)                                                    // List
		userRouter.Get(usernamePattern, s.usersService.Read(chiUsernameFetcher))                    // Read
		userRouter.Delete(usernamePattern, s.usersService.Delete(chiUserIDFetcher))                 // Delete
		userRouter.With(s.usersService.UserInputContextMiddleware).Post("/", s.usersService.Create) // Create
		// userRouter.With(s.usersService.UserInputContextMiddleware).Put(sr, s.usersService.Update)   // Update
	})

	router.Route("/oauth2", func(oauth2Router chi.Router) {
		oauth2Router.
			With(s.Oauth2ClientInfoMiddleware).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
				if err := s.oauth2Handler.HandleAuthorizeRequest(res, req); err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
				}
			})

		oauth2Router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2Handler.HandleTokenRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)

			}
		})
	})

	router.With(s.OauthTokenAuthenticationMiddleware).Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {

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

	s.router = router
	s.server.Handler = router
}
