package frontend

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// these tests are not parallelized because goment has race conditions.

func Test_mustGoment(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ts := uint64(time.Now().Unix())

		assert.NotNil(t, mustGoment(ts))
	})
}

func Test_relativeTime(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ts := uint64(time.Now().Unix())

		assert.NotEmpty(t, relativeTime(ts))
	})
}

func Test_relativeTimeFromPtr(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ts := uint64(time.Now().Unix())

		assert.NotEmpty(t, relativeTimeFromPtr(&ts))
	})
}
