package zerolog

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

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

		assert.NotNil(t, NewLogger())
	})
}

func Test_logger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.WithName(t.Name()))
	})
}

func Test_logger_SetLevel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.SetLevel(logging.ErrorLevel)
	})
}

func Test_logger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.SetRequestIDFunc(func(*http.Request) string {
			return ""
		})
	})
}

func Test_logger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.Info(t.Name())
	})
}

func Test_logger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.Debug(t.Name())
	})
}

func Test_logger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.Error(errors.New("blah"), t.Name())
	})
}

func Test_logger_Printf(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		l.Printf(t.Name())
	})
}

func Test_logger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.Clone())
	})
}

func Test_logger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.WithValue("name", t.Name()))
	})
}

func Test_logger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.WithValues(map[string]interface{}{"name": t.Name()}))
	})
}

func Test_logger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.WithError(errors.New("blah")))
	})
}

func Test_logger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger().(*logger)
		l.requestIDFunc = func(r *http.Request) string {
			return t.Name()
		}

		u, err := url.ParseRequestURI("https://todo.verygoodsoftwarenotvirus.ru?things=stuff")
		require.NoError(t, err)

		assert.NotNil(t, l.WithRequest(&http.Request{
			URL: u,
		}))
	})
}

func Test_logger_WithResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()

		assert.NotNil(t, l.WithResponse(&http.Response{}))
	})
}
