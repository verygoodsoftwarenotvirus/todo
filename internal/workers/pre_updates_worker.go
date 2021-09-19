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

// PreUpdateMessage represents an event that asks a worker to update data to the datastore.
type PreUpdateMessage struct {
	MessageType       string      `json:"messageType"`
	Item              *types.Item `json:"item"`
	AttributeToUserID string      `json:"userID"`
}

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
func (w *PreUpdatesWorker) HandleMessage(message []byte) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *PreUpdateMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributeToUserID)

	w.logger.WithValue("message_type", msg.MessageType).WithValue("item", msg.Item).Debug("message read")

	switch msg.MessageType {
	case "item":
		if err := w.dataManager.UpdateItem(ctx, msg.Item); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postUpdatesPublisher != nil {
			if err := w.postUpdatesPublisher.Publish(ctx, msg.Item); err != nil {
				w.logger.Error(err, "publishing to post-updates topic")
			}
		}
	}

	return nil
}
