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
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildAPIClientCreationEventEntry(&types.APIClient{}, exampleUserID),
		},
		"BuildAPIClientArchiveEventEntry": {
			expectedEventType: audit.APIClientArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.AccountAssignmentKey,
				audit.APIClientAssignmentKey,
			},
			actual: audit.BuildAPIClientArchiveEventEntry(exampleAccountID, exampleAPIClientDatabaseID, exampleUserID),
		},
	}

	runEventBuilderTests(T, tests)
}
