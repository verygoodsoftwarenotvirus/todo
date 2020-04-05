package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_opencensusCounter_Increment(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		ct, err := ProvideUnitCounter("v", "description")
		c, typOk := ct.(*opencensusCounter)
		require.NoError(t, err)
		require.True(t, typOk)

		assert.Equal(t, c.actualCount, uint64(0))

		c.Increment(ctx)
		assert.Equal(t, c.actualCount, uint64(1))
	})
}

func Test_opencensusCounter_IncrementBy(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		ct, err := ProvideUnitCounter("v", "description")
		c, typOk := ct.(*opencensusCounter)
		require.NoError(t, err)
		require.True(t, typOk)

		assert.Equal(t, c.actualCount, uint64(0))

		c.IncrementBy(ctx, 666)
		assert.Equal(t, c.actualCount, uint64(666))
	})
}

func Test_opencensusCounter_setCountTo(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		ct, err := ProvideUnitCounter("v", "description")
		c, typOk := ct.(*opencensusCounter)
		require.NoError(t, err)
		require.True(t, typOk)

		assert.Equal(t, c.actualCount, uint64(0))

		c.IncrementBy(ctx, 123)
		assert.Equal(t, c.actualCount, uint64(123))

		c.Decrement(ctx)
		expected := uint64(666)

		c.setCountTo(ctx, expected)
		actual := c.actualCount

		assert.Equal(t, expected, actual)
	})
}

func Test_opencensusCounter_Decrement(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		ct, err := ProvideUnitCounter("v", "description")
		c, typOk := ct.(*opencensusCounter)
		require.NoError(t, err)
		require.True(t, typOk)

		assert.Equal(t, c.actualCount, uint64(0))

		c.Increment(ctx)
		assert.Equal(t, c.actualCount, uint64(1))

		c.Decrement(ctx)
		assert.Equal(t, c.actualCount, uint64(0))
	})
}

func TestProvideUnitCounterProvider(t *testing.T) {
	t.Parallel()

	// obligatory
	assert.NotNil(t, ProvideUnitCounterProvider())
}
