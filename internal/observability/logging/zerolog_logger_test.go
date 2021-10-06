package logging

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildZerologger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, buildZerologger())
	})
}

func TestNewLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewZerologLogger())
	})
}

func Test_zerologLogger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.WithName(t.Name()))
	})
}

func Test_zerologLogger_SetLevel(T *testing.T) {
	T.Parallel()

	T.Run("Info", func(t *testing.T) {
		t.Parallel()

		NewZerologLogger().SetLevel(InfoLevel)
	})

	T.Run("Debug", func(t *testing.T) {
		t.Parallel()

		NewZerologLogger().SetLevel(DebugLevel)
	})

	T.Run("Error", func(t *testing.T) {
		t.Parallel()

		NewZerologLogger().SetLevel(ErrorLevel)
	})

	T.Run("Warn", func(t *testing.T) {
		t.Parallel()

		NewZerologLogger().SetLevel(WarnLevel)
	})
}

func Test_zerologLogger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		l.SetRequestIDFunc(func(*http.Request) string {
			return ""
		})
	})
}

func Test_zerologLogger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		l.Info(t.Name())
	})
}

func Test_zerologLogger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		l.Debug(t.Name())
	})
}

func Test_zerologLogger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		l.Error(errors.New("blah"), t.Name())
	})
}

func Test_zerologLogger_Printf(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		l.Printf(t.Name())
	})
}

func Test_zerologLogger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.Clone())
	})
}

func Test_zerologLogger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.WithValue("name", t.Name()))
	})
}

func Test_zerologLogger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.WithValues(map[string]interface{}{"name": t.Name()}))
	})
}

func Test_zerologLogger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.WithError(errors.New("blah")))
	})
}

func Test_zerologLogger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l, ok := NewZerologLogger().(*zerologLogger)
		require.True(t, ok)

		l.requestIDFunc = func(*http.Request) string {
			return t.Name()
		}

		u, err := url.ParseRequestURI("https://todo.verygoodsoftwarenotvirus.ru?things=stuff")
		require.NoError(t, err)

		assert.NotNil(t, l.WithRequest(&http.Request{
			URL: u,
		}))
	})
}

func Test_zerologLogger_WithResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZerologLogger()

		assert.NotNil(t, l.WithResponse(&http.Response{}))
	})
}
