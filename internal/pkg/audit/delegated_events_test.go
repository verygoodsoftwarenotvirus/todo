package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAPIClientDatabaseID uint64 = 123
)

func TestBuildAPIClientCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAPIClientCreationEventEntry(&types.APIClient{}, exampleUserID))
}

func TestBuildAPIClientArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAPIClientArchiveEventEntry(exampleAccountID, exampleAPIClientDatabaseID, exampleUserID))
}
