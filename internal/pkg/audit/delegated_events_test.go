package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	exampleDelegatedClientID uint64 = 123
)

func TestDelegatedEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]*eventBuilderTest{
		"BuildDelegatedClientCreationEventEntry": {
			expectedEventType: audit.DelegatedClientCreationEvent,
			expectedContextKeys: []string{
				audit.DelegatedClientAssignmentKey,
				audit.CreationAssignmentKey,
			},
			actual: audit.BuildDelegatedClientCreationEventEntry(&types.DelegatedClient{}),
		},
		"BuildDelegatedClientArchiveEventEntry": {
			expectedEventType: audit.DelegatedClientArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.DelegatedClientAssignmentKey,
			},
			actual: audit.BuildDelegatedClientArchiveEventEntry(exampleUserID, exampleDelegatedClientID),
		},
	}

	runEventBuilderTests(T, tests)
}
