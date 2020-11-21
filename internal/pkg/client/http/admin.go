package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
)

const (
	adminBasePath = "_admin_"
)

// BuildBanUserRequest builds a request to ban a user.
func (c *V1Client) BuildBanUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildBanUserRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		adminBasePath,
		usersBasePath,
		strconv.FormatUint(userID, 10),
		"ban",
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BanUser executes a request to ban a user.
func (c *V1Client) BanUser(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "BanUser")
	defer span.End()

	req, err := c.BuildBanUserRequest(ctx, userID)
	if err != nil {
		return fmt.Errorf("error building user ban request: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	c.closeResponseBody(res)

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("erroneous response code when banning user: %d", res.StatusCode)
	}

	return nil
}
