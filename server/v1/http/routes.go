package httpserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// func bareMiddlewareBlueprint(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
// 		next.ServeHTTP(res, req)
// 	})
// }

func (s *Server) setupRouter(frontendFilesPath string, metricsHandler metrics.Handler) {
	router := chi.NewRouter()

	router.Use(
		middleware.RequestID,
		middleware.Timeout(maxTimeout),
		s.loggingMiddleware,
	)

	// all middleware must be defined before routes on a mux

	// define client side rendered asset httpServer
	pwd, _ := os.Getwd()
	filesDir := filepath.Join(pwd, frontendFilesPath)
	fs := http.StripPrefix("/", http.FileServer(http.Dir(filesDir)))
	router.Get("/*", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/listyourfrontendhistoryroutesrouteshere":
			req.URL.Path = "/"
		}

		fs.ServeHTTP(res, req)
	}))

	// health check
	router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) { res.WriteHeader(http.StatusOK) })
	router.Handle("/metrics", metricsHandler)

	router.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.usersService.UserLoginInputMiddleware).Post("/login", s.authService.Login)
		userRouter.With(s.authService.CookieAuthenticationMiddleware).Post("/logout", s.authService.Logout)

		userIDPattern := fmt.Sprintf(`/{%s:[0-9_\-]+}`, users.URIParamKey)

		userRouter.Get("/", s.usersService.List)                                             // List
		userRouter.With(s.usersService.UserInputMiddleware).Post("/", s.usersService.Create) // Create
		userRouter.Get(userIDPattern, s.usersService.Read)                                   // Read
		userRouter.Delete(userIDPattern, s.usersService.Delete)                              // Delete

		userRouter.With(
			s.authService.CookieAuthenticationMiddleware,
			s.usersService.TOTPSecretRefreshInputMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecret)

		userRouter.With(
			s.authService.CookieAuthenticationMiddleware,
			s.usersService.PasswordUpdateInputMiddleware,
		).Post("/password/new", s.usersService.UpdatePassword)
	})

	router.Route("/oauth2", func(oauth2Router chi.Router) {
		oauth2Router.With(
			s.authService.CookieAuthenticationMiddleware,
			s.oauth2ClientsService.CreationInputMiddleware,
		).Post("/client", s.oauth2ClientsService.Create) // Create

		oauth2Router.With(s.oauth2ClientsService.OAuth2ClientInfoMiddleware).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
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
		Route("/api", func(apiRouter chi.Router) {
			apiRouter.Route("/v1", func(v1Router chi.Router) {

				// Items
				v1Router.Route("/items", func(itemsRouter chi.Router) {
					sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
					itemsRouter.With(s.itemsService.CreationInputMiddleware).Post("/", s.itemsService.Create) // Create
					itemsRouter.Get(sr, s.itemsService.Read)                                                  // Read
					itemsRouter.With(s.itemsService.UpdateInputMiddleware).Put(sr, s.itemsService.Update)     // Update
					itemsRouter.Delete(sr, s.itemsService.Delete)                                             // Delete
					itemsRouter.Get("/", s.itemsService.List)                                                 // List
				})

				// OAuth2 Clients
				v1Router.Route("/oauth2/clients", func(clientRouter chi.Router) {
					sr := fmt.Sprintf(`/{%s:[0-9]+}`, oauth2clients.URIParamKey)
					// Create is not bound to an OAuth2 authentication token
					// Update not supported for OAuth2 clients.
					clientRouter.Get(sr, s.oauth2ClientsService.Read)      // Read
					clientRouter.Delete(sr, s.oauth2ClientsService.Delete) // Delete
					clientRouter.Get("/", s.oauth2ClientsService.List)     // List
				})

			})

		})

	s.router = router
}
