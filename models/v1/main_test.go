package models

import (
	"testing"
)

func TestErrorResponse_Error(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = (&ErrorResponse{}).Error()
	})
}
