package testutil

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ContextMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance, in that the resulting function will panic if anything other than a context.Context.
func ContextMatcher(keys ...types.ContextKey) func(context.Context) bool {
	return func(ctx context.Context) bool {
		for key := range keys {
			if x := ctx.Value(key); x == nil {
				return false
			}
		}

		return true
	}
}

// RequestMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance, in that the resulting function will panic if anything other than a *http.Request.
func RequestMatcher() func(*http.Request) bool {
	return func(req *http.Request) bool {
		return true
	}
}

// ResponseWriterMatcher is a matcher for use with testify/mock's MatchBy function. It provides some level of type
// safety reassurance, in that the resulting function will panic if anything other than a http.ResponseWriter.
func ResponseWriterMatcher() func(http.ResponseWriter) bool {
	return func(res http.ResponseWriter) bool {
		return true
	}
}
