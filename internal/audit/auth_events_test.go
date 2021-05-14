package audit_test

import (
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"

	"github.com/stretchr/testify/assert"
)

func TestBuildCycleCookieSecretEvent(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildCycleCookieSecretEvent(exampleUserID))
}

func TestBuildSuccessfulLoginEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildSuccessfulLoginEventEntry(exampleUserID))
}

func TestBannedUserLoginAttemptEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildBannedUserLoginAttemptEventEntry(exampleUserID))
}

func TestBuildUnsuccessfulLoginBadPasswordEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleUserID))
}

func TestBuildUnsuccessfulLoginBad2FATokenEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleUserID))
}

func TestBuildLogoutEventEntry(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, audit.BuildLogoutEventEntry(exampleUserID))
}
