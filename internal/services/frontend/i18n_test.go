package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_getSimpleLocalizedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		assert.Equal(t, ":)", s.service.getSimpleLocalizedString("testing.translation"))
	})
}

func Test_provideLocalizer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, provideLocalizer())
	})
}
