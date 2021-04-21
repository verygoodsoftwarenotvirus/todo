package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

var (
	urlToUse       string
	parsedURLToUse *url.URL
)

func init() {
	ctx := context.Background()

	parsedURLToUse = testutils.DetermineServiceURL()
	urlToUse = parsedURLToUse.String()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutils.EnsureServerIsUp(ctx, urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
