package workers

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// PreUpdatesWorker updates data from the pending updates topic to the database.
type PreUpdatesWorker struct {
	logger               logging.Logger
	tracer               tracing.Tracer
	encoder              encoding.ClientEncoder
	postUpdatesPublisher publishers.Publisher
	dataManager          database.DataManager
}

// ProvidePreUpdatesWorker provides a PreUpdatesWorker.
func ProvidePreUpdatesWorker(logger logging.Logger, dataManager database.DataManager, postUpdatesPublisher publishers.Publisher) *PreUpdatesWorker {
	const name = "pre_updates"

	return &PreUpdatesWorker{
		logger:               logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:               tracing.NewTracer(name),
		encoder:              encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postUpdatesPublisher: postUpdatesPublisher,
		dataManager:          dataManager,
	}
}

// HandleMessage handles a pending update.
func (w *PreUpdatesWorker) HandleMessage(ctx context.Context, message []byte) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	var msg *types.PreUpdateMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributableToUserID)

	w.logger.WithValue("data_type", msg.DataType).Debug("message read")

	switch msg.DataType {
	case types.ItemDataType:
		if err := w.dataManager.UpdateItem(ctx, msg.Item); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postUpdatesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				Item:                    msg.Item,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}
			if err := w.postUpdatesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-updates topic")
			}
		}
	case types.UserMembershipDataType, types.WebhookDataType:
		break
	}

	return nil
}
