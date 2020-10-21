package httpserver

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	idReplacementRegex = regexp.MustCompile(`[^(v|oauth)]\\d+`)
)

func formatSpanNameForRequest(req *http.Request) string {
	return fmt.Sprintf(
		"%s %s",
		req.Method,
		idReplacementRegex.ReplaceAllString(req.URL.Path, "/{id}"),
	)
}

var doMotLog = map[string]struct{}{
	"/metrics": {},
	"/build/":  {},
	"/assets/": {},
}

func buildLoggingMiddleware(logger logging.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ww := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(ww, req)

			shouldLog := true
			for route := range doMotLog {
				if !strings.HasPrefix(req.URL.Path, route) || req.URL.Path == route {
					shouldLog = false
					break
				}
			}

			if shouldLog {
				logger.WithRequest(req).WithValues(map[string]interface{}{
					"status":        ww.Status(),
					"bytes_written": ww.BytesWritten(),
					"elapsed":       time.Since(start),
				}).Debug("responded to request")
			}
		})
	}
}

/*
func httpsRedirectMiddleware(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		if strings.EqualFold(req.URL.Scheme, "http") {
			res.Header().Set("Connection", "close")
			req.URL.Scheme = "https"
			http.Redirect(res, req, req.URL.String(), http.StatusMovedPermanently)
		} else {
			next.ServeHTTP(res, req)
		}
	}
	return http.HandlerFunc(fn)
}
*/
