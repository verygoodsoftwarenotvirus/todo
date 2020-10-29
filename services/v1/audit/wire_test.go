package audit

import (
	"testing"
)

func TestProvideAuditLogEntryDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideAuditLogEntryDataServer(buildTestService())
	})
}
