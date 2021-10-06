package tracing

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartCustomSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		StartCustomSpan(ctx, "blah")
	})

	T.Run("with nil ctx", func(t *testing.T) {
		t.Parallel()

		//nolint:staticcheck // ignore SA1012 in tests
		StartCustomSpan(nil, "blah")
	})
}

func TestStartSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		StartSpan(ctx)
	})

	T.Run("with nil ctx", func(t *testing.T) {
		t.Parallel()

		//nolint:staticcheck // ignore SA1012 in tests
		StartSpan(nil)
	})
}

func TestFormatSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u, err := url.ParseRequestURI("https://todo.verygoodsoftwarenotvirus.ru")
		require.NoError(t, err)

		FormatSpan(t.Name(), &http.Request{URL: u})
	})
}
