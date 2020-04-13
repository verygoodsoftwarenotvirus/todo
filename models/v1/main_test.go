package models

import (
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v5"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

func TestErrorResponse_Error(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = (&ErrorResponse{}).Error()
	})
}
