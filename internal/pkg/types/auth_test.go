package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionInfo_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := &SessionInfo{
			UserID:      123,
			UserIsAdmin: true,
		}

		assert.NotEmpty(t, x.ToBytes())
	})
}
