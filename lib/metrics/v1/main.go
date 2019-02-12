package metrics

import (
	"net/http"
)

type (

	// Namespace is a string alias for dependency injection's sake
	Namespace string

	// InstrumentationHandler is the Handler that provides instrumentation details at the root of the server mux
	InstrumentationHandler http.Handler

	// InstrumentationHandlerProvider is a function that builds an InstrumentationHandler
	InstrumentationHandlerProvider func(http.Handler) InstrumentationHandler

	// Handler is the Handler that provides metrics data to scraping services
	Handler http.Handler
)
