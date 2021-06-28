package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuildUserAddedToAccountEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserAddedToAccountEventEntry(exampleAdminUserID, &types.AddUserToAccountInput{Reason: t.Name()}))
}

func TestBuildUserRemovedFromAccountEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserRemovedFromAccountEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID, "blah blah"))
}

func TestBuildUserMarkedAccountAsDefaultEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserMarkedAccountAsDefaultEventEntry(exampleAdminUserID, exampleUserID, exampleAccountID))
}

func TestBuildModifyUserPermissionsEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildModifyUserPermissionsEventEntry(exampleUserID, exampleAccountID, exampleAdminUserID, []string{t.Name()}, t.Name()))
}

func TestBuildTransferAccountOwnershipEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildTransferAccountOwnershipEventEntry(exampleAccountID, exampleAdminUserID, fakes.BuildFakeTransferAccountOwnershipInput()))
}
