package httpserver

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/heptiolabs/healthcheck"
)

// func bareMiddlewareBlueprint(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
// 		next.ServeHTTP(res, req)
// 	})
// }

const (
	registrationRoute = `/users`
	loginRoute        = `/users/login`
	numericIDPattern  = `/{%s:[0-9]+}`
	oauth2IDPattern   = `/{%s:[0-9_\-]+}`
)

func (s *Server) setupRouter(frontendConfig config.FrontendSettings, metricsHandler metrics.Handler) {
	router := chi.NewRouter()

	// Basic CORS, for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	ch := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	router.Use(
		middleware.RequestID,
		middleware.Timeout(maxTimeout),
		s.loggingMiddleware,
		ch.Handler,
	)

	// all middleware must be defined before routes on a mux

	router.Route("/_meta_", func(metaRouter chi.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		router.Handle("/metrics", metricsHandler)
	}

	// Frontend routes
	if frontendConfig.StaticFilesDirectory != "" {
		staticFileServer, err := s.frontendService.StaticDir(frontendConfig.StaticFilesDirectory)
		if err != nil {
			s.logger.Error(err, "establishing static file server")
		}
		router.Get("/*", staticFileServer)
	}

	for route, handler := range s.frontendService.Routes() {
		router.Get(route, handler)
	}

	router.With(
		s.authService.AuthenticationMiddleware(true),
		s.authService.AdminMiddleware,
	).Route("/admin", func(adminRouter chi.Router) {
		adminRouter.Post("/cycle_cookie_secret", s.authService.CycleSecret)
	})

	router.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.authService.UserLoginInputMiddleware).
			Post("/login", s.authService.Login)
		userRouter.With(s.authService.CookieAuthenticationMiddleware).
			Post("/logout", s.authService.Logout)

		userIDPattern := fmt.Sprintf(oauth2IDPattern, users.URIParamKey)

		userRouter.Get("/", s.usersService.List) // List
		userRouter.With(s.usersService.UserInputMiddleware).
			Post("/", s.usersService.Create) // Create
		userRouter.Get(userIDPattern, s.usersService.Read)      // Read
		userRouter.Delete(userIDPattern, s.usersService.Delete) // Delete

		userRouter.With(
			s.authService.CookieAuthenticationMiddleware,
			s.usersService.TOTPSecretRefreshInputMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecret)

		userRouter.With(
			s.authService.CookieAuthenticationMiddleware,
			s.usersService.PasswordUpdateInputMiddleware,
		).Put("/password/new", s.usersService.UpdatePassword)
	})

	router.Route("/oauth2", func(oauth2Router chi.Router) {
		oauth2Router.With(
			s.authService.CookieAuthenticationMiddleware,
			s.oauth2ClientsService.CreationInputMiddleware,
		).Post("/client", s.oauth2ClientsService.Create) // Create

		oauth2Router.With(s.oauth2ClientsService.OAuth2ClientInfoMiddleware).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
				s.logger.WithRequest(req).Debug("oauth2 authorize route hit")
				if err := s.oauth2ClientsService.HandleAuthorizeRequest(res, req); err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
				}
			})

		oauth2Router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2ClientsService.HandleTokenRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
		})
	})

	router.
		With(s.authService.AuthenticationMiddleware(true)).
		Route("/api/v1", func(v1Router chi.Router) {

			// Items
			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := fmt.Sprintf(numericIDPattern, items.URIParamKey)
				itemsRouter.With(s.itemsService.CreationInputMiddleware).
					Post("/", s.itemsService.Create) // Create
				itemsRouter.Get(sr, s.itemsService.Read) // Read
				itemsRouter.With(s.itemsService.UpdateInputMiddleware).
					Put(sr, s.itemsService.Update) // Update
				itemsRouter.Delete(sr, s.itemsService.Delete) // Delete
				itemsRouter.Get("/", s.itemsService.List)     // List
			})

			// Webhooks
			v1Router.Route("/webhooks", func(webhookRouter chi.Router) {
				sr := fmt.Sprintf(numericIDPattern, webhooks.URIParamKey)
				webhookRouter.With(s.webhooksService.CreationInputMiddleware).
					Post("/", s.webhooksService.Create) // Create
				webhookRouter.Get(sr, s.webhooksService.Read) // Read
				webhookRouter.With(s.webhooksService.UpdateInputMiddleware).
					Put(sr, s.webhooksService.Update) // Update
				webhookRouter.Delete(sr, s.webhooksService.Delete) // Delete
				webhookRouter.Get("/", s.webhooksService.List)     // List
			})

			// OAuth2 Clients
			v1Router.Route("/oauth2/clients", func(clientRouter chi.Router) {
				sr := fmt.Sprintf(numericIDPattern, oauth2clients.URIParamKey)
				// Create is not bound to an OAuth2 authentication token
				// Update not supported for OAuth2 clients.
				clientRouter.Get(sr, s.oauth2ClientsService.Read)      // Read
				clientRouter.Delete(sr, s.oauth2ClientsService.Delete) // Delete
				clientRouter.Get("/", s.oauth2ClientsService.List)     // List
			})

		})

	s.router = router
}
