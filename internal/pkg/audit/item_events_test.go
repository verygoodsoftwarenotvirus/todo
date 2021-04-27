package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleItemID uint64 = 123
)

func TestBuildItemCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemCreationEventEntry(&types.Item{}, exampleAccountID))
}

func TestBuildItemUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemUpdateEventEntry(exampleUserID, exampleItemID, exampleAccountID, nil))
}

func TestBuildItemArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemArchiveEventEntry(exampleUserID, exampleItemID, exampleAccountID))
}
