package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"

	"github.com/stretchr/testify/assert"
)

func TestBuildUserBanEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserBanEventEntry(exampleUserID, exampleUserID, "reason"))
}
