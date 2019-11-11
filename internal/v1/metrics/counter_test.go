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
		ct, err := ProvideUnitCounter("counter", "description")
		c := ct.(*opencensusCounter)

		require.NoError(t, err)
		assert.Equal(t, c.actualCount, uint64(0))

		c.Increment(context.Background())
		assert.Equal(t, c.actualCount, uint64(1))
	})
}

func Test_opencensusCounter_IncrementBy(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ct, err := ProvideUnitCounter("counter", "description")
		c := ct.(*opencensusCounter)

		require.NoError(t, err)
		assert.Equal(t, c.actualCount, uint64(0))

		c.IncrementBy(context.Background(), 666)
		assert.Equal(t, c.actualCount, uint64(666))
	})
}

func Test_opencensusCounter_Decrement(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ct, err := ProvideUnitCounter("counter", "description")
		c := ct.(*opencensusCounter)

		require.NoError(t, err)
		assert.Equal(t, c.actualCount, uint64(0))

		c.Increment(context.Background())
		assert.Equal(t, c.actualCount, uint64(1))

		c.Decrement(context.Background())
		assert.Equal(t, c.actualCount, uint64(0))
	})
}

func TestProvideUnitCounterProvider(T *testing.T) {
	T.Parallel()

	// obligatory
	assert.NotNil(T, ProvideUnitCounterProvider())
}
