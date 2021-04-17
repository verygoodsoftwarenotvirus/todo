package audit_test

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
)

func TestBuildUserAddedToAccountEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserAddedToAccountEventEntry(exampleAdminUserID, &types.AddUserToAccountInput{}))
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

	assert.NotNil(t, audit.BuildModifyUserPermissionsEventEntry(exampleUserID, exampleAccountID, exampleAdminUserID, testutil.BuildNoUserPerms(), ""))
}

func TestBuildTransferAccountOwnershipEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildTransferAccountOwnershipEventEntry(exampleAccountID, exampleAdminUserID, fakes.BuildFakeTransferAccountOwnershipInput()))
}
