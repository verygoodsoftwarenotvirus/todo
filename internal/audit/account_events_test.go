package audit_test

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAccountID uint64 = 123
)

func TestBuildAccountCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountCreationEventEntry(&types.Account{}, exampleUserID))
}

func TestBuildAccountUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountUpdateEventEntry(exampleUserID, exampleAccountID, exampleUserID, nil))
}

func TestBuildAccountArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountArchiveEventEntry(exampleUserID, exampleAccountID, exampleUserID))
}
