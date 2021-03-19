package requests

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	adminBasePath = "_admin_"
)

// BuildUserReputationUpdateInputRequest builds a request to change a user's reputation.
func (b *Builder) BuildUserReputationUpdateInputRequest(ctx context.Context, input *types.UserReputationUpdateInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.UserIDKey, input.TargetUserID)

	if err := input.Validate(ctx); err != nil {
		return nil, prepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, adminBasePath, usersBasePath, "status")

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}
