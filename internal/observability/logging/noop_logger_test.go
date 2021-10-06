package logging

import (
	"errors"
	"testing"
)

func TestNoopLogger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).Info(t.Name())
	})
}

func TestNoopLogger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).Debug(t.Name())
	})
}

func TestNoopLogger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).Error(errors.New(t.Name()), t.Name())
	})
}

func TestNoopLogger_Printf(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).Printf(t.Name())
	})
}

func TestNoopLogger_SetLevel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).SetLevel(InfoLevel)
	})
}

func TestNoopLogger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).SetRequestIDFunc(nil)
	})
}

func TestNoopLogger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithName(t.Name())
	})
}

func TestNoopLogger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).Clone()
	})
}

func TestNoopLogger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithValues(nil)
	})
}

func TestNoopLogger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithValue(t.Name(), t.Name())
	})
}

func TestNoopLogger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithRequest(nil)
	})
}

func TestNoopLogger_WithResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithResponse(nil)
	})
}

func TestNoopLogger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		(&noopLogger{}).WithError(errors.New(t.Name()))
	})
}
