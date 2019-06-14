package models

import (
	"testing"
)

func TestNewDataExport(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		NewDataExport()
	})
}
