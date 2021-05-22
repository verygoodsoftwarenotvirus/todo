package types

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChangeActiveAccountInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &ChangeActiveAccountInput{
			AccountID: 123,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestPASETOCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &PASETOCreationInput{
			ClientID:    t.Name(),
			RequestTime: time.Now().Unix(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestSessionContextData_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &SessionContextData{}

		assert.NotEmpty(t, x.ToBytes())
	})
}
