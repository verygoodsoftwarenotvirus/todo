package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleItemID = "123"
)

func TestBuildItemCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemCreationEventEntry(&types.Item{}, exampleUserID))
}

func TestBuildItemUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemUpdateEventEntry(exampleUserID, exampleItemID, exampleAccountID, nil))
}

func TestBuildItemArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildItemArchiveEventEntry(exampleUserID, exampleAccountID, exampleItemID))
}
