package testutils

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"

	"github.com/stretchr/testify/mock"
)

// ContextMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a context.Context.
var ContextMatcher interface{} = mock.MatchedBy(func(context.Context) bool {
	return true
})

// HTTPRequestMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a *http.Request.
var HTTPRequestMatcher interface{} = mock.MatchedBy(func(*http.Request) bool {
	return true
})

// HTTPResponseWriterMatcher is a matcher for the http.ResponseWriter interface. It provides some level of type
// safety reassurance over mock.Anything, in that the resulting function will panic if anything other than
// a http.ResponseWriter.
var HTTPResponseWriterMatcher interface{} = mock.MatchedBy(func(http.ResponseWriter) bool {
	return true
})

// PreWriteMessageMatcher matches the workers.PreWriteMessage type.
func PreWriteMessageMatcher(*workers.PreWriteMessage) bool { return true }

// PreUpdateMessageMatcher matches the workers.PreUpdateMessage type.
func PreUpdateMessageMatcher(*workers.PreUpdateMessage) bool { return true }

// PreArchiveMessageMatcher matches the workers.PreArchiveMessage type.
func PreArchiveMessageMatcher(*workers.PreArchiveMessage) bool { return true }
