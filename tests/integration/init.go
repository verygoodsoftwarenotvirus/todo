package integration

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

const (
	debug         = false
	timeout       = 5 * time.Second
	nonexistentID = 999999999
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
