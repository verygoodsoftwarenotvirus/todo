package types

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

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		_ = (&ErrorResponse{}).Error()
	})
}
