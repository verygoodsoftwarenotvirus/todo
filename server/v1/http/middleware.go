package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

func (s *Server) tracingMiddleware(next http.Handler) http.Handler {
	return nethttp.Middleware(
		s.tracer,
		next,
		nethttp.MWComponentName("todo-httpServer"),
		nethttp.MWSpanObserver(func(span opentracing.Span, req *http.Request) {
			span.SetTag("http.method", req.Method)
			span.SetTag("http.uri", req.URL.EscapedPath())
		}),
		nethttp.OperationNameFunc(func(req *http.Request) string {
			return fmt.Sprintf("%s %s", req.Method, req.URL.Path)
		}),
	)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ww := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

		start := time.Now()
		defer func() {
			s.logger.WithValues(map[string]interface{}{
				"status":        ww.Status(),
				"bytes_written": ww.BytesWritten(),
				"elapsed":       time.Since(start),
			})
		}()

		next.ServeHTTP(ww, req)
	})
}
