package metrics

import (
	"testing"
	"time"
)

func TestRecordRuntimeStats(T *testing.T) {
	T.Parallel()

	// this is sort of an obligatory test for coverage's sake

	d := time.Second
	sf := RecordRuntimeStats(d / 5)
	time.Sleep(d)
	sf()
}
