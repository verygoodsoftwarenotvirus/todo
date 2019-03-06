package metrics

import (
	"net/http"
)

type (

	// Namespace is a string alias for dependency injection's sake
	Namespace string

	// Middleware is our middleware
	Middleware func(http.Handler) http.Handler

	// InstrumentationHandler is an obligatory alias
	InstrumentationHandler http.Handler

	// InstrumentationHandlerProvider is a function that builds an InstrumentationHandler
	InstrumentationHandlerProvider func(http.Handler) InstrumentationHandler

	// Handler is the Handler that provides metrics data to scraping services
	Handler http.Handler

	// HandlerInstrumentationFunc blah
	HandlerInstrumentationFunc func(http.HandlerFunc) http.HandlerFunc
)
