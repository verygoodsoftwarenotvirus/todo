package requests

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

const (
	websocketsBasePath = "websockets"
)

// BuildSubscribeToDataChangesWebsocketURL builds a URL for subscribing to a websocket to receive DataChangeMessages.
func (b *Builder) BuildSubscribeToDataChangesWebsocketURL(ctx context.Context) string {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildWebsocketURL(
		ctx,
		websocketsBasePath,
		"data_changes",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return uri
}
