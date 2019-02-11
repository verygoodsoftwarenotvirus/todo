package server

import (
	"log"
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	// KeyMethod is a key representing method
	KeyMethod, _ = tag.NewKey("method")

	// mLatencyMs represents The latency in milliseconds
	mLatencyMs = stats.Float64("repl/latency", "The latency in milliseconds per REPL loop", "ms")

	// MLineLengths counts/groups the lengths of lines read in.
	MLineLengths = stats.Int64("repl/line_lengths", "The distribution of line lengths", "By")

	// LatencyView is a latency view
	LatencyView = &view.View{
		Name:        "demo/latency",
		Measure:     mLatencyMs,
		Description: "The distribution of the latencies",

		// Latency in buckets:
		// [>=0ms, >=25ms, >=50ms, >=75ms, >=100ms, >=200ms, >=400ms, >=600ms, >=800ms, >=1s, >=2s, >=4s, >=6s]
		Aggregation: view.Distribution(0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000),
		TagKeys:     []tag.Key{KeyMethod},
	}
)

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ochttp.WithRouteTag(next, req.URL.Path).ServeHTTP(res, req)
		stats.Record(
			req.Context(),
			mLatencyMs.M(float64(time.Since(start).Nanoseconds())),
		)
	})
}

func initViews() {
	if err := view.Register(LatencyView, nil); err != nil {
		log.Fatalf("Failed to register the views: %v", err)
	}
}
