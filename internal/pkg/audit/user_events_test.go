package audit_test

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"testing"
)

const (
	exampleUserID uint64 = 123
)

func TestUserEventBuilders(T *testing.T) {
	T.Parallel()

	tests := map[string]eventBuilderTest{
		"BuildUserCreationEventEntry": {
			expectedEventType: audit.UserCreationEvent,
			expectedContextKeys: []string{
				audit.CreationAssignmentKey,
				audit.UserAssignmentKey,
			},
			actual: audit.BuildUserCreationEventEntry(&types.User{}),
		},
		"BuildUserVerifyTwoFactorSecretEventEntry": {
			expectedEventType: audit.UserVerifyTwoFactorSecretEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildUserVerifyTwoFactorSecretEventEntry(exampleUserID),
		},
		"BuildUserUpdateTwoFactorSecretEventEntry": {
			expectedEventType: audit.UserUpdateTwoFactorSecretEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildUserUpdateTwoFactorSecretEventEntry(exampleUserID),
		},
		"BuildUserUpdatePasswordEventEntry": {
			expectedEventType: audit.UserUpdatePasswordEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
			},
			actual: audit.BuildUserUpdatePasswordEventEntry(exampleUserID),
		},
		"BuildUserArchiveEventEntry": {
			expectedEventType: audit.UserArchiveEvent,
			expectedContextKeys: []string{
				audit.ActorAssignmentKey,
				audit.UserAssignmentKey,
			},
			actual: audit.BuildUserArchiveEventEntry(exampleUserID),
		},
	}

	testEventBuilders(T, tests)
}
