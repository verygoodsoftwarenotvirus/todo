package http

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/errs"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
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

	logger := c.logger.WithValue(keys.AccountIDKey, input.TargetUserID)
	tracing.AttachAccountIDToSpan(span, input.TargetUserID)

	if err := input.Validate(ctx); err != nil {
		return errs.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildUserReputationUpdateInputRequest(ctx, input)
	if err != nil {
		return errs.PrepareError(err, logger, span, "building user reputation update request")
	}

	res, err := c.fetchResponseToRequest(ctx, c.authedClient, req)
	if err != nil {
		return errs.PrepareError(err, logger, span, "updating user reputation")
	}

	c.closeResponseBody(ctx, res)

	if err = errorFromResponse(res); err != nil {
		return errs.PrepareError(err, logger, span, "invalid response status")
	}

	return nil
}
