package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

type eventBuilderTest struct {
	actual              *types.AuditLogEntryCreationInput
	expectedEventType   string
	expectedContextKeys []string
}

func runEventBuilderTests(T *testing.T, tests map[string]*eventBuilderTest) {
	T.Helper()

	for name, test := range tests {
		tName := name
		tTest := test

		T.Run(tName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tTest.expectedEventType, tTest.actual.EventType, "expected event type to be %v, was %v in test %q", tTest.expectedEventType, tTest.actual.EventType, tName)
			for k := range tTest.actual.Context {
				assert.Contains(t, tTest.expectedContextKeys, k, "expected %q to be present in request context in test %q", k, tName)
			}
		})
	}
}
