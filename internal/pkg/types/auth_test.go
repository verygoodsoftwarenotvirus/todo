package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestContext_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RequestContext{Requester: RequesterInfo{ID: 123}}

		assert.NotEmpty(t, x.ToBytes())
	})
}
