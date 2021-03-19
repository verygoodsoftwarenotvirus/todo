package types

import (
	"fmt"
	"testing"

	"github.com/RussellLuo/validating/v2"
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

		// much as we'd like to use testutil.InvalidRawURL here, it causes a cyclical import :'(
		assert.NotNil(t, x.Validate(fmt.Sprintf("%s://verygoodsoftwarenotvirus.ru", string(byte(127)))))
	})

	T.Run("invalid value", func(t *testing.T) {
		t.Parallel()
		x := &urlValidator{}

		assert.NotNil(t, x.Validate(validating.F("arbitrary", 123)))
	})
}
