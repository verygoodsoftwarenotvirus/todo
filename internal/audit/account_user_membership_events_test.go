package audit_test

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
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
