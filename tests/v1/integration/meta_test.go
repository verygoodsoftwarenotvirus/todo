package integration

import (
	"os"
	"testing"
	"time"
)

func TestHoldOnForever(T *testing.T) {
	T.Parallel()

	if os.Getenv("WAIT_FOR_COVERAGE") == "yes" {
		time.Sleep(time.Hour * 24 * 365)
	}
}
