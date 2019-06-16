package httpserver

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/middleware"
)

var (
	idReplacementRegex = regexp.MustCompile(`[^(v|oauth)]\d+`)
)

func formatSpanNameForRequest(req *http.Request) string {
	return fmt.Sprintf(
		"%s %s",
		req.Method,
		idReplacementRegex.ReplaceAllString(req.URL.Path, `/{id}`),
	)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ww := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

		start := time.Now()
		next.ServeHTTP(ww, req)

		s.logger.WithRequest(req).WithValues(map[string]interface{}{
			"status":        ww.Status(),
			"bytes_written": ww.BytesWritten(),
			"elapsed":       time.Since(start),
		}).Debug("request received")
	})
}
