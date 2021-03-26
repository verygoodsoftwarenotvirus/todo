package testutil

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ContextMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a context.Context.
func ContextMatcher(ctx context.Context) bool {
	return true
}

// AuditLogEntryCreationInputMatcher is a matcher for use with testify/mock's MatchBy function.
func AuditLogEntryCreationInputMatcher(eventType string) func(*types.AuditLogEntryCreationInput) bool {
	return func(input *types.AuditLogEntryCreationInput) bool {
		return input.EventType == eventType
	}
}

// RequestMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a *http.Request.
func RequestMatcher() func(*http.Request) bool {
	return func(req *http.Request) bool {
		return true
	}
}

// ResponseWriterMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a http.ResponseWriter.
func ResponseWriterMatcher() func(http.ResponseWriter) bool {
	return func(res http.ResponseWriter) bool {
		return true
	}
}
