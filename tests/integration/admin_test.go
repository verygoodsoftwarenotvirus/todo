package integration

import (
	"log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func (s *TestSuite) TestAdmin_Returns404WhenModifyingUserReputation() {
	s.runForEachClientExcept("should not be possible to ban a user that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			input := fakes.BuildFakeUserReputationUpdateInput()
			input.TargetUserID = nonexistentID

			// Ban user.
			assert.Error(t, testClients.admin.UpdateUserReputation(ctx, input))
		}
	})
}

func (s *TestSuite) TestAdmin_BanningUsers() {
	s.runForEachClientExcept("should be possible to ban users", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			var (
				user       *types.User
				userClient *httpclient.Client
			)

			switch testClients.authType {
			case cookieAuthType:
				user, _, userClient, _ = createUserAndClientForTest(ctx, t)
			case pasetoAuthType:
				user, _, _, userClient = createUserAndClientForTest(ctx, t)
			default:
				log.Panicf("invalid auth type: %q", testClients.authType)
			}

			// Assert that user can access service
			_, initialCheckErr := userClient.GetAPIClients(ctx, nil)
			require.NoError(t, initialCheckErr)

			input := &types.UserReputationUpdateInput{
				TargetUserID:  user.ID,
				NewReputation: types.BannedUserAccountStatus,
				Reason:        "testing",
			}

			assert.NoError(t, testClients.admin.UpdateUserReputation(ctx, input))

			// Assert user can no longer access service
			_, subsequentCheckErr := userClient.GetAPIClients(ctx, nil)
			assert.Error(t, subsequentCheckErr)

			// Clean up.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, user.ID))
		}
	})
}
