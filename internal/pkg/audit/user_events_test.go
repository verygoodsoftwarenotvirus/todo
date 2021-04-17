package audit_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
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

func TestBuildUserArchiveEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUserArchiveEventEntry(exampleUserID))
}
