package audit_test

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAccountSubscriptionPlanID uint64 = 123
)

func TestBuildAccountSubscriptionPlanCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountSubscriptionPlanCreationEventEntry(&types.AccountSubscriptionPlan{}))
}
func TestBuildAccountSubscriptionPlanUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountSubscriptionPlanUpdateEventEntry(exampleUserID, exampleAccountSubscriptionPlanID, nil))
}
func TestBuildAccountSubscriptionPlanArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildAccountSubscriptionPlanArchiveEventEntry(exampleUserID, exampleAccountSubscriptionPlanID))
}
