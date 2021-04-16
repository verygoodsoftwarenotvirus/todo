package random

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardSecretGenerator_GenerateEncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s := NewGenerator(nil)
		value, err := s.GenerateBase64EncodedString(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Greater(t, len(value), exampleLength)
		assert.NoError(t, err)
	})
}

func TestStandardSecretGenerator_GenerateRawBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s := NewGenerator(nil)
		value, err := s.GenerateRawBytes(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Equal(t, len(value), exampleLength)
		assert.NoError(t, err)
	})
}
