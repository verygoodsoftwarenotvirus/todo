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

// PreWriteMessage represents an event that asks a worker to write data to the datastore.
type PreWriteMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// PreWritesWorker writes data from the pending writes topic to the database.
type PreWritesWorker struct {
	logger              logging.Logger
	tracer              tracing.Tracer
	encoder             encoding.ClientEncoder
	postWritesPublisher publishers.Publisher
	dataManager         database.DataManager
}

// ProvidePreWritesWorker provides a PreWritesWorker.
func ProvidePreWritesWorker(logger logging.Logger, dataManager database.DataManager, postWritesPublisher publishers.Publisher) *PreWritesWorker {
	const name = "pre_writes"

	return &PreWritesWorker{
		logger:              logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:              tracing.NewTracer(name),
		encoder:             encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postWritesPublisher: postWritesPublisher,
		dataManager:         dataManager,
	}
}

// HandleMessage handles a pending write.
func (w *PreWritesWorker) HandleMessage(message []byte) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *PreWriteMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributableToUserID)

	w.logger.WithValue("message_type", msg.MessageType).WithValue("item", msg.Item).Debug("message read")

	switch msg.MessageType {
	case "item":
		_, err := w.dataManager.CreateItem(ctx, msg.Item)
		if err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postWritesPublisher != nil {
			if err = w.postWritesPublisher.Publish(ctx, msg.Item); err != nil {
				w.logger.Error(err, "publishing to post-writes topic")
			}
		}
	}

	return nil
}
