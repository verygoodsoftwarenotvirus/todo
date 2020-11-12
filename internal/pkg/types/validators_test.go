package types

import (
	"fmt"
	"testing"

	v "github.com/RussellLuo/validating/v2"
	"github.com/stretchr/testify/assert"
)

func Test_minimumStringLengthValidator_Validate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		x := &minimumStringLengthValidator{minLength: 1}

		assert.Nil(t, x.Validate(v.F("arbitrary", "blah")))
	})

	T.Run("unhappy path", func(t *testing.T) {
		t.Parallel()
		x := &minimumStringLengthValidator{minLength: 1}

		assert.NotNil(t, x.Validate(v.F("arbitrary", "")))
	})
}

func Test_minimumStringSliceLengthValidator_Validate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		x := &minimumStringSliceLengthValidator{minLength: 1}
		y := []string{"blah"}

		assert.Nil(t, x.Validate(v.F("arbitrary", &y)))
	})

	T.Run("unhappy path", func(t *testing.T) {
		t.Parallel()
		x := &minimumStringSliceLengthValidator{minLength: 1}
		y := []string{}

		assert.NotNil(t, x.Validate(v.F("arbitrary", &y)))
	})
}

func Test_urlValidator_Validate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		x := &urlValidator{}

		assert.Nil(t, x.Validate(v.F("arbitrary", "https://verygoodsoftwarenotvirus.ru")))
	})

	T.Run("unhappy path", func(t *testing.T) {
		t.Parallel()
		x := &urlValidator{}

		assert.NotNil(t, x.Validate(v.F("arbitrary", fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(127))))))
	})
}
