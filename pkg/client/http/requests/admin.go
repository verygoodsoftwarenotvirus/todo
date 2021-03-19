package requests

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	adminBasePath = "_admin_"
)

// BuildUserReputationUpdateInputRequest builds a request to ban a user.
func (b *Builder) BuildUserReputationUpdateInputRequest(ctx context.Context, input *types.UserReputationUpdateInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		b.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := b.BuildURL(ctx, nil, adminBasePath, usersBasePath, "status")

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}
