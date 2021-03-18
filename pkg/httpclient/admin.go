package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// UpdateUserReputation updates a user's reputation.
func (c *Client) UpdateUserReputation(ctx context.Context, input *types.UserReputationUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		tracing.AttachErrorToSpan(span, validationErr)
		return fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.requestBuilder.BuildUserReputationUpdateInputRequest(ctx, input)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building user account status update request: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("executing request: %w", err)
	}

	c.closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("erroneous response code when banning user: %d", res.StatusCode)
	}

	return nil
}
