package frontend

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	exampleTestKey = "testing"
)

func Test_anyToString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "thing", anyToString("thing"))
	})
}

func Test_stringToBool(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"true"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToBool(exampleForm, exampleTestKey)

		assert.True(t, actual)
	})

	T.Run("with error", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToBool(exampleForm, exampleTestKey)

		assert.False(t, actual)
	})
}

func Test_stringToFloat32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := float32(123.45)

		exampleForm := url.Values{
			exampleTestKey: []string{"123.45"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToFloat32(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToFloat32(exampleForm, exampleTestKey)
	})
}

func Test_stringToFloat64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := float64(123.45)

		exampleForm := url.Values{
			exampleTestKey: []string{"123.45"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToFloat64(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToFloat64(exampleForm, exampleTestKey)
	})
}

func Test_stringToInt(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToInt(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToInt(exampleForm, exampleTestKey)
	})
}

func Test_stringToInt16(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int16(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToInt16(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToInt16(exampleForm, exampleTestKey)
	})
}

func Test_stringToInt32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int32(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToInt32(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToInt32(exampleForm, exampleTestKey)
	})
}

func Test_stringToInt64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int64(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToInt64(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToInt64(exampleForm, exampleTestKey)
	})
}

func Test_stringToInt8(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int8(123)

		exampleForm := url.Values{
			exampleTestKey: []string{"123"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToInt8(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToInt8(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToBool(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := true

		exampleForm := url.Values{
			exampleTestKey: []string{"true"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToBool(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToBool(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToFloat32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := float32(123.45)

		exampleForm := url.Values{
			exampleTestKey: []string{"123.45"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToFloat32(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToFloat32(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToFloat64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := float64(123.45)

		exampleForm := url.Values{
			exampleTestKey: []string{"123.45"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToFloat64(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToFloat64(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToInt(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToInt(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToInt(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToInt16(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int16(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToInt16(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToInt16(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToInt32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int32(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToInt32(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToInt32(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToInt64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int64(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToInt64(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToInt64(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToInt8(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int8(123)

		exampleForm := url.Values{
			exampleTestKey: []string{"123"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToInt8(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToInt8(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := "hello"

		exampleForm := url.Values{
			exampleTestKey: []string{"hello"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToString(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToString(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToUint(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToUint(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToUint(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToUint16(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint16(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToUint16(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToUint16(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToUint32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint32(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToUint32(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToUint32(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToUint64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint64(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToUint64(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToUint64(exampleForm, exampleTestKey)
	})
}

func Test_stringToPointerToUint8(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint8(123)

		exampleForm := url.Values{
			exampleTestKey: []string{"123"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToPointerToUint8(exampleForm, exampleTestKey)

		assert.Equal(t, &expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToPointerToUint8(exampleForm, exampleTestKey)
	})
}

func Test_stringToUint(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToUint(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToUint(exampleForm, exampleTestKey)
	})
}

func Test_stringToUint16(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint16(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToUint16(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToUint16(exampleForm, exampleTestKey)
	})
}

func Test_stringToUint32(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint32(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToUint32(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToUint32(exampleForm, exampleTestKey)
	})
}

func Test_stringToUint64(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint64(12345)

		exampleForm := url.Values{
			exampleTestKey: []string{"12345"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToUint64(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToUint64(exampleForm, exampleTestKey)
	})
}

func Test_stringToUint8(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint8(123)

		exampleForm := url.Values{
			exampleTestKey: []string{"123"},
		}

		s := buildTestHelper(t)
		actual := s.service.stringToUint8(exampleForm, exampleTestKey)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		exampleForm := url.Values{
			exampleTestKey: []string{"lol"},
		}

		s := buildTestHelper(t)
		s.service.stringToUint8(exampleForm, exampleTestKey)
	})
}
