package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionInfo_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RequestContext{User: UserRequestContext{ID: 123}}

		assert.NotEmpty(t, x.ToBytes())
	})
}
