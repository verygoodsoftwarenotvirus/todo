package testutil

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

// ContextMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a context.Context.
var ContextMatcher interface{} = mock.MatchedBy(func(context.Context) bool {
	return true
})

// RequestMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a *http.Request.
var RequestMatcher interface{} = mock.MatchedBy(func(*http.Request) bool {
	return true
})

// ResponseWriterMatcher is a matcher for the http.ResponseWriter interface. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a http.ResponseWriter.
var ResponseWriterMatcher interface{} = mock.MatchedBy(func(http.ResponseWriter) bool {
	return true
})

// AuditLogEntryCreationInputMatcher is a matcher for use with testify/mock's MatchBy function.
func AuditLogEntryCreationInputMatcher(eventType string) func(*types.AuditLogEntryCreationInput) bool {
	return func(input *types.AuditLogEntryCreationInput) bool {
		return input.EventType == eventType
	}
}
