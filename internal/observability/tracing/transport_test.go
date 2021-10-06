package tracing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildWrappedTransport(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, BuildTracedHTTPTransport(time.Minute))
	})
}
