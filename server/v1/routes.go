package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gcontext "github.com/gorilla/context"
)

func (s *Server) buildRouteCtx(key models.ContextKey, x interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if err := json.NewDecoder(req.Body).Decode(x); err != nil {
				s.logger.Errorf("error encountered decoding request body: %v", err)
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(res, req.WithContext(context.WithValue(req.Context(), key, x)))
		})
	}
}

func (s *Server) setupRoutes() {
	s.router = chi.NewRouter()
	s.router.Use(
		gcontext.ClearHandler,
		middleware.RequestID,
		middleware.DefaultLogger,
		middleware.Timeout(maxTimeout),
	)

	if s.DebugMode {
		s.router.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
		s.router.Get("/_debug_/stats", s.stats)
	}

	s.router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) { res.WriteHeader(http.StatusOK) })

	s.router.Route("/users", func(userRouter chi.Router) {
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

		usernamePattern := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", users.URIParamKey)

		userRouter.Get("/", s.usersService.List)                                                    // List
		userRouter.Get(usernamePattern, s.usersService.Read(chiUsernameFetcher))                    // Read
		userRouter.Delete(usernamePattern, s.usersService.Delete(chiUsernameFetcher))               // Delete
		userRouter.With(s.usersService.UserInputContextMiddleware).Post("/", s.usersService.Create) // Create
		// userRouter.With(s.usersService.UserInputContextMiddleware).Put(sr, s.usersService.Update)   // Update
	})

	s.router.Route("/oauth2", func(oauth2Router chi.Router) {
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

	s.router.
		With(s.OauthTokenAuthenticationMiddleware).
		Route("/api", func(apiRouter chi.Router) {
			apiRouter.Route("/v1", func(v1Router chi.Router) {

				v1Router.Route("/items", func(itemsRouter chi.Router) {
					sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
					itemsRouter.Get("/", s.itemsService.List)                                   // List
					itemsRouter.Get("/count", s.itemsService.Count)                             // Count
					itemsRouter.Get(sr, s.itemsService.BuildReadHandler(chiItemIDFetcher))      // Read
					itemsRouter.Delete(sr, s.itemsService.BuildDeleteHandler(chiItemIDFetcher)) // Delete
					itemsRouter.With(s.itemsService.ItemInputMiddleware).
						Put(sr, s.itemsService.BuildUpdateHandler(chiItemIDFetcher)) // Update
					itemsRouter.With(s.itemsService.ItemInputMiddleware).
						Post("/", s.itemsService.Create) // Create
				})

				v1Router.Route("/oauth2", func(oauth2Router chi.Router) {
					oauth2Router.Route("/clients", func(clientRouter chi.Router) {
						sr := fmt.Sprintf("/{%s}", oauth2ClientIDURIParamKey)
						clientRouter.Get("/", s.ListOauth2Clients)    // List
						clientRouter.Get(sr, s.ReadOauth2Client)      // Read
						clientRouter.Delete(sr, s.DeleteOauth2Client) // Delete
						clientRouter.
							With(s.buildRouteCtx(oauth2ClientIDKey, new(models.Oauth2ClientUpdateInput))).
							Put(sr, s.UpdateOauth2Client) // Update
						clientRouter.
							With(s.buildRouteCtx(oauth2ClientIDKey, new(models.Oauth2ClientCreationInput))).
							Post("/", s.CreateOauth2Client) // Create
					})
				})

			})
		})
}
