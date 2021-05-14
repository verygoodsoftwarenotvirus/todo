package audit_test

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAdminUserID uint64 = 321
	exampleUserID      uint64 = 123
)

func TestBuildUserCreationEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserCreationEventEntry(exampleUserID))
}

func TestBuildUserVerifyTwoFactorSecretEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserVerifyTwoFactorSecretEventEntry(exampleUserID))
}

func TestBuildUserUpdateTwoFactorSecretEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserUpdateTwoFactorSecretEventEntry(exampleUserID))
}

func TestBuildUserUpdatePasswordEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserUpdatePasswordEventEntry(exampleUserID))
}

func TestBuildUserUpdateEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserUpdateEventEntry(exampleUserID, nil))
}

func TestBuildUserArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserArchiveEventEntry(exampleUserID))
}
