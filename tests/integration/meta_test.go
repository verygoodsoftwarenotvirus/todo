package integration

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestHoldOnForever(T *testing.T) {
	T.Parallel()

	if os.Getenv("WAIT_FOR_COVERAGE") == "yes" {
		// snooze for a year.
		time.Sleep(time.Hour * 24 * 365)
	}
}

var _ suite.WithStats = (*TestSuite)(nil)

func (s *TestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 48

	if stats.Passed() {
		s.Equal(totalExpectedTestCount, len(stats.TestStats), "expected total number of tests run to equal %d, but it was %d", totalExpectedTestCount, len(stats.TestStats))
	}
}
