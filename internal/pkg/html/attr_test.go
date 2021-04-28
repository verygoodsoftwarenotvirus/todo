package html

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestAttribute(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := Attr{
			"class": "testing",
			"name":  t.Name(),
		}
		actual := Attribute(
			WithClass("testing"),
			WithValue("name", t.Name()),
		)

		assert.Equal(t, expected, actual)
	})
}

func TestAttr_Modify(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := Attr{
			"class": "testing",
			"name":  t.Name(),
		}
		actual := Attribute(
			WithClass("testing"),
			WithValue("name", t.Name()),
		)

		assert.Equal(t, expected, actual)

		expected["things"] = "stuff"

		x := actual.Modify(
			WithValue("things", "stuff"),
		)

		assert.Equal(t, expected, actual)
		assert.Equal(t, x, actual)
	})
}

func TestWithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		key := "name"
		actual := Attribute(
			WithValue(key, t.Name()),
		)

		assert.Equal(t, t.Name(), actual[key])
	})
}

func TestWithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		key1 := "things"
		key2 := "stuff"
		expected := map[string]string{
			key1: t.Name(),
			key2: strings.ToLower(t.Name()),
		}

		actual := Attribute(
			WithValues(expected),
		)

		for k, v := range expected {
			assert.Equal(t, v, actual[k], "key %q had an unexpected attr value", k)
		}
	})
}

func TestWithClass(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		class := t.Name()
		actual := Attribute(
			WithClass(class),
		)

		assert.Equal(t, class, actual["class"])
	})
}

func TestWithClasses(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := "things and stuff"
		actual := Attribute(
			WithClasses("things", "and", "stuff"),
		)

		assert.Equal(t, expected, actual["class"])
	})
}
