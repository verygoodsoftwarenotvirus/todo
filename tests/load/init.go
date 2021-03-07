package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

var (
	urlToUse string
	cookie   *http.Cookie
)

func init() {
	ctx := context.Background()
	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	u, err := utils.CreateServiceUser(ctx, urlToUse, "")
	if err != nil {
		logger.Fatal(err)
	}

	cookie, err = utils.GetLoginCookie(ctx, urlToUse, u)
	if err != nil {
		logger.Fatal(err)
	}

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
