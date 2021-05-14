package integration

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	zerolog "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

const (
	debug                = true
	nonexistentID uint64 = math.MaxUint32
)

var (
	urlToUse       string
	parsedURLToUse *url.URL

	premadeAdminUser = &types.User{
		ID:              1,
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Username:        "exampleUser",
		HashedPassword:  "integration-tests-are-cool",
	}
)

func init() {
	ctx, span := tracing.StartSpan(context.Background())
	defer span.End()

	parsedURLToUse = testutil.DetermineServiceURL()
	urlToUse = parsedURLToUse.String()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
