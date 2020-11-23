package integration

import (
	"context"
	"testing"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/require"
)

func createUserAndClientForTest(ctx context.Context, t *testing.T) (*types.User, *client.V1Client) {
	t.Helper()

	user, err := testutil.CreateObligatoryUser(ctx, urlToUse, "", debug)
	require.NoError(t, err)

	oa2Client, err := testutil.CreateObligatoryClient(ctx, urlToUse, user)
	require.NoError(t, err)

	return user, initializeClient(oa2Client)
}
