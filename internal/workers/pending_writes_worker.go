package workers

import (
	"context"
	"encoding/json"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// PendingWriteMessage represents an event that asks a worker to write data to the datastore.
type PendingWriteMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// PendingWritesWorker writes data from the pending writes topic to the database.
type PendingWritesWorker struct {
	logger              logging.Logger
	tracer              tracing.Tracer
	afterWritesProducer messagequeue.Producer
	errorsProducer      messagequeue.Producer
	dataManager         database.DataManager
}

// ProvidePendingWritesWorker provides a PendingWritesWorker.
func ProvidePendingWritesWorker(logger logging.Logger, dataManager database.DataManager) *PendingWritesWorker {
	name := "pending_writes"

	return &PendingWritesWorker{
		logger:      logging.EnsureLogger(logger).WithName(name),
		tracer:      tracing.NewTracer(name),
		dataManager: dataManager,
	}
}

// HandleMessage handles a pending write.
func (w *PendingWritesWorker) HandleMessage(message []byte) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *PendingWriteMessage

	if err := json.Unmarshal(message, &msg); err != nil {
		if w.errorsProducer != nil {
			if errPublishErr := w.errorsProducer.Publish(ctx, err); errPublishErr != nil {
				w.logger.Error(errPublishErr, "publishing error to errors topic")
			}
		}

		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	w.logger.WithValue("message_type", msg.MessageType).WithValue("item", msg.Item).Debug("message read")

	switch msg.MessageType {
	case "item":
		_, err := w.dataManager.CreateItem(ctx, msg.Item)
		if err != nil {
			if w.errorsProducer != nil {
				if errPublishErr := w.errorsProducer.Publish(ctx, err); errPublishErr != nil {
					w.logger.Error(errPublishErr, "publishing error to errors topic")
				}
			}

			return observability.PrepareError(err, w.logger, span, "creating item")
		}
	}

	return nil
}
