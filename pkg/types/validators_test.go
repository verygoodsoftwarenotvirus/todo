package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_urlValidator_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		x := &urlValidator{}

		assert.Nil(t, x.Validate("https://verygoodsoftwarenotvirus.ru"))
	})

	T.Run("unhappy path", func(t *testing.T) {
		t.Parallel()
		x := &urlValidator{}

		// much as we'd like to use testutils.InvalidRawURL here, it causes a cyclical import :'(
		assert.NotNil(t, x.Validate(fmt.Sprintf("%s://verygoodsoftwarenotvirus.ru", string(byte(127)))))
	})
}

func Test_stringDurationValidator_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &stringDurationValidator{maxDuration: time.Hour}

		assert.NoError(t, x.Validate(time.Minute.String()))
	})

	T.Run("invalid value", func(t *testing.T) {
		t.Parallel()

		x := &stringDurationValidator{maxDuration: time.Hour}

		assert.Error(t, x.Validate(1234))
	})

	T.Run("invalid format", func(t *testing.T) {
		t.Parallel()

		x := &stringDurationValidator{maxDuration: time.Hour}

		assert.Error(t, x.Validate("fake lol"))
	})

	T.Run("too large a max duration", func(t *testing.T) {
		t.Parallel()

		x := &stringDurationValidator{maxDuration: time.Hour}

		assert.Error(t, x.Validate((2400 * time.Hour).String()))
	})
}
