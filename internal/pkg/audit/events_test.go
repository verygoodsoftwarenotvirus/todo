package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

type eventBuilderTest struct {
	expectedEventType   string
	expectedContextKeys []string
	actual              *types.AuditLogEntryCreationInput
}

func runEventBuilderTests(T *testing.T, tests map[string]*eventBuilderTest) {
	T.Helper()

	for name, test := range tests {
		name := name
		test := test
		T.Run(name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			assert.Equal(t, test.expectedEventType, test.actual.EventType, "expected event type to be %v, was %v", test.expectedEventType, test.actual.EventType)
			for k := range test.actual.Context {
				assert.Contains(t, test.expectedContextKeys, k, "expected %q to be present in request context in test %q", k, name)
			}
		})
	}
}
