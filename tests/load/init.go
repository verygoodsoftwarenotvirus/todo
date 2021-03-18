package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

var (
	urlToUse       string
	parsedURLToUse *url.URL
)

func init() {
	ctx := context.Background()

	parsedURLToUse = testutil.DetermineServiceURL()
	urlToUse = parsedURLToUse.String()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
