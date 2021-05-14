package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditLogContext_Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &AuditLogContext{}

		actual, err := x.Value()
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestAuditLogContext_Scan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &AuditLogContext{}

		assert.NoError(t, x.Scan([]byte("{}")))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		x := &AuditLogContext{}

		assert.Error(t, x.Scan(t.Name()))
	})
}
