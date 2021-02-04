package chi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

func buildRouterForTest() routing.Router {
	return NewRouter(logging.NewNonOperationalLogger())
}

func TestNewRouter(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		assert.NotNil(t, NewRouter(logging.NewNonOperationalLogger()))
	})
}

func Test_buildChiMux(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		assert.NotNil(t, buildChiMux(
			logging.NewNonOperationalLogger(),
			tracing.NewTracer("test"),
		))
	})
}

func Test_buildLoggingMiddleware(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		middleware := buildLoggingMiddleware(logging.NewNonOperationalLogger(), tracing.NewTracer("blah"))

		assert.NotNil(t, middleware)

		hf := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

		req, res := httptest.NewRequest(http.MethodPost, "/nil", nil), httptest.NewRecorder()

		middleware(hf).ServeHTTP(res, req)
	})
}

func Test_convertMiddleware(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		middleware := convertMiddleware(func(http.Handler) http.Handler { return nil })
		assert.NotNil(t, middleware)
	})
}

func Test_router_AddRoute(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		methods := []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}

		for _, method := range methods {
			assert.NoError(t, r.AddRoute(method, "/path", nil))
		}
	})

	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		assert.Error(t, r.AddRoute("fart", "/path", nil))
	})
}

func Test_router_Connect(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Connect("/test", nil)
	})
}

func Test_router_Delete(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Delete("/test", nil)
	})
}

func Test_router_Get(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Get("/test", nil)
	})
}

func Test_router_Handle(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Handle("/test", nil)
	})
}

func Test_router_HandleFunc(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.HandleFunc("/test", nil)
	})
}

func Test_router_Handler(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		assert.NotNil(t, r.Handler())
	})
}

func Test_router_Head(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Head("/test", nil)
	})
}

func Test_router_LogRoutes(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		assert.NoError(t, r.AddRoute(http.MethodGet, "/path", nil))

		r.LogRoutes()
	})
}

func Test_router_Options(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Options("/test", nil)
	})
}

func Test_router_Patch(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Patch("/test", nil)
	})
}

func Test_router_Post(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Post("/test", nil)
	})
}

func Test_router_Put(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Put("/thing", nil)
	})
}

func Test_router_Route(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		assert.NotNil(t, r.Route("/test", func(routing.Router) {}))
	})
}

func Test_router_Trace(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		r.Trace("/test", nil)
	})
}

func Test_router_WithMiddleware1(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouterForTest()

		assert.NotNil(t, r.WithMiddleware())
	})
}

func Test_router_clone(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		r := buildRouter(nil, nil)

		assert.NotNil(t, r.clone())
	})
}
