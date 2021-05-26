package frontend

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

var urlToUse string

func init() {
	u := testutil.DetermineServiceURL()
	urlToUse = u.String()

	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})
	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(context.Background(), urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
