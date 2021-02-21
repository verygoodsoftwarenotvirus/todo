package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleAPIClientDatabaseID uint64 = 123
)

func TestAPIClientEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildAPIClientCreationEventEntry": {
			expectedEventType: audit.APIClientCreationEvent,
			expectedContextKeys: []string{
				audit.APIClientAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildAPIClientCreationEventEntry(&types.APIClient{}),
		},
		"BuildAPIClientArchiveEventEntry": {
			expectedEventType: audit.APIClientArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.APIClientAssignmentKey,
			},
			actual: audit.BuildAPIClientArchiveEventEntry(exampleUserID, exampleAPIClientDatabaseID),
		},
	}

	runEventBuilderTests(T, tests)
}
