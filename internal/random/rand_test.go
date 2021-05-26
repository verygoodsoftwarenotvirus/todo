package random

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type erroneousReader struct{}

func (r *erroneousReader) Read(p []byte) (n int, err error) {
	return -1, errors.New("blah")
}

func TestGenerateBase32EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		actual, err := GenerateBase32EncodedString(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestGenerateBase64EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		actual, err := GenerateBase64EncodedString(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestGenerateRawBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		actual, err := GenerateRawBytes(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestStandardSecretGenerator_GenerateBase32EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s := NewGenerator(nil)
		value, err := s.GenerateBase32EncodedString(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Greater(t, len(value), exampleLength)
		assert.NoError(t, err)
	})

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s, ok := NewGenerator(nil).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateBase32EncodedString(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
	})
}

func TestStandardSecretGenerator_GenerateBase64EncodedString(T *testing.T) {
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

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s, ok := NewGenerator(nil).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateBase64EncodedString(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
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

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleLength := 123

		s, ok := NewGenerator(nil).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateRawBytes(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
	})
}
