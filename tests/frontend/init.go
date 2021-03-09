package frontend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

var urlToUse string

const (
	seleniumHubAddr    = "http://selenium-hub:4444/wd/hub"
	seleniumServerWait = 10 * time.Second
)

func init() {
	urlToUse = testutil.DetermineServiceURL()

	logger := zerolog.NewLogger()
	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(context.Background(), urlToUse)

	// NOTE: this is sad, but also the only thing that consistently works
	// see above for my vain attempts at a real solution to this problem.
	time.Sleep(seleniumServerWait)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
