package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	adminBasePath = "_admin_"
)

// BuildAccountStatusUpdateInputRequest builds a request to ban a user.
func (c *Client) BuildAccountStatusUpdateInputRequest(ctx context.Context, input *types.UserReputationUpdateInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, adminBasePath, usersBasePath, "status")

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// UpdateAccountStatus executes a request to ban a user.
func (c *Client) UpdateAccountStatus(ctx context.Context, input *types.UserReputationUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildAccountStatusUpdateInputRequest(ctx, input)
	if err != nil {
		return fmt.Errorf("building user account status update request: %w", err)
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
