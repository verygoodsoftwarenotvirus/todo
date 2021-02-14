package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionInfo_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := &SessionInfo{UserID: 123}

		assert.NotEmpty(t, x.ToBytes())
	})
}
