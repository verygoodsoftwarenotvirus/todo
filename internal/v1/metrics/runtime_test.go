package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecordRuntimeStats(T *testing.T) {
	T.Parallel()

	// this is sort of an obligatory test for coverage's sake.

	d := time.Second
	sf := RecordRuntimeStats(d / 5)
	time.Sleep(d)
	sf()
}

func TestRegisterDefaultViews(t *testing.T) {
	t.Parallel()

	// obligatory
	require.NoError(t, RegisterDefaultViews())
}
