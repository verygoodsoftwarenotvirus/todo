package panicking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProductionPanicker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewProductionPanicker())
	})
}

func Test_stdLibPanicker_Panic(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := NewProductionPanicker()

		defer func() {
			assert.NotNil(t, recover(), "expected panic to occur")
		}()

		p.Panic("blah")
	})
}

func Test_stdLibPanicker_Panicf(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := NewProductionPanicker()

		defer func() {
			assert.NotNil(t, recover(), "expected panic to occur")
		}()

		p.Panicf("blah")
	})
}
