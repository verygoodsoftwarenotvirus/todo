package integration

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHoldOnForever(T *testing.T) {
	T.Parallel()

	if os.Getenv("WAIT_FOR_COVERAGE") == "yes" {
		// snooze for a year
		time.Sleep(time.Hour * 24 * 365)
	}
}

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, i)
}
