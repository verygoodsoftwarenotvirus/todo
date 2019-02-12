package metrics

import (
	"net/http"
)

// Namespace is a string alias for dependency injection's sake
type Namespace string

// Collector is our abstracted interface for a metrics collection
// service like Prometheus
type Collector interface {
	MetricsHandler() http.Handler
}
