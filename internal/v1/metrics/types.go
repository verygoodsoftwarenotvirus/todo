package metrics

import (
	"context"
	"net/http"
)

type (
	// Namespace is a string alias for dependency injection's sake
	Namespace string

	// CounterName is a string alias for dependency injection's sake
	CounterName string

	// SpanFormatter formats the name of a span given a request
	SpanFormatter func(*http.Request) string

	// InstrumentationHandler is an obligatory alias
	InstrumentationHandler http.Handler

	// Handler is the Handler that provides metrics data to scraping services
	Handler http.Handler

	// HandlerInstrumentationFunc blah
	HandlerInstrumentationFunc func(http.HandlerFunc) http.HandlerFunc

	// UnitCounter describes a counting interface for things like total user counts
	// Meant to handle integers exclusively
	UnitCounter interface {
		Increment(ctx context.Context)
		IncrementBy(ctx context.Context, val uint64)
		Decrement(ctx context.Context)
	}

	// UnitCounterProvider is a function that provides a UnitCounter and an error
	UnitCounterProvider func(counterName CounterName, description string) (UnitCounter, error)
)
