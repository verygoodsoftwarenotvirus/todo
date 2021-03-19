package integration

import (
	"fmt"
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestAdminUserManagement() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to ban a user that does not exist via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			input := fakes.BuildFakeAccountStatusUpdateInput()
			input.TargetAccountID = nonexistentID

			// Ban user.
			assert.Error(t, testClients.admin.UpdateUserReputation(ctx, input))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to ban users via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			var (
				user       *types.User
				userClient *httpclient.Client
			)

			switch authType {
			case cookieAuthType:
				user, _, userClient, _ = createUserAndClientForTest(ctx, t)
			case pasetoAuthType:
				user, _, _, userClient = createUserAndClientForTest(ctx, t)
			default:
				log.Panicf("invalid auth type: %q", authType)
			}

			// Assert that user can access service
			_, initialCheckErr := userClient.GetAPIClients(ctx, nil)
			require.NoError(t, initialCheckErr)

			input := &types.UserReputationUpdateInput{
				TargetAccountID: user.ID,
				NewReputation:   types.BannedAccountStatus,
				Reason:          "testing",
			}

			assert.NoError(t, testClients.admin.UpdateUserReputation(ctx, input))

			// Assert user can no longer access service
			_, subsequentCheckErr := userClient.GetAPIClients(ctx, nil)
			assert.Error(t, subsequentCheckErr)

			// Clean up.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, user.ID))
		})
	}
}
