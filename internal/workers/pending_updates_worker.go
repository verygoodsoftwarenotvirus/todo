package workers

import (
	"context"
	"encoding/json"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/nsqio/go-nsq"
)

// PendingUpdateMessage represents an event that asks a worker to write data to the datastore.
type PendingUpdateMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// PendingUpdatesWorker writes data from the pending writes topic to the database.
type PendingUpdatesWorker struct {
	logger              logging.Logger
	tracer              tracing.Tracer
	afterWritesProducer events.Producer
	errorsProducer      events.Producer
	dataManager         database.DataManager
}

// ProvidePendingUpdatesWorker provides a PendingUpdatesWorker.
func ProvidePendingUpdatesWorker(logger logging.Logger, dataManager database.DataManager, afterWritesProducer, errorsProducer events.Producer) *PendingUpdatesWorker {
	name := "pending_updates"

	return &PendingUpdatesWorker{
		logger:              logging.EnsureLogger(logger).WithName(name),
		tracer:              tracing.NewTracer(name),
		afterWritesProducer: afterWritesProducer,
		errorsProducer:      errorsProducer,
		dataManager:         dataManager,
	}
}

// HandleMessage handles a pending write.
func (w *PendingUpdatesWorker) HandleMessage(message *nsq.Message) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *PendingUpdateMessage

	if err := json.Unmarshal(message.Body, &msg); err != nil {
		message.Touch()

		if errPublishErr := w.errorsProducer.Publish(ctx, err); errPublishErr != nil {
			w.logger.Error(errPublishErr, "publishing error to errors topic")
		}

		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	w.logger.WithValue("message_type", msg.MessageType).WithValue("item", msg.Item).Debug("message read")

	switch msg.MessageType {
	case "item":
		_, err := w.dataManager.CreateItem(ctx, msg.Item)
		if err != nil {
			message.Touch()

			if errPublishErr := w.errorsProducer.Publish(ctx, err); errPublishErr != nil {
				w.logger.Error(errPublishErr, "publishing error to errors topic")
			}

			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		message.Finish()

		if err = w.afterWritesProducer.Publish(ctx, msg.Item); err != nil {
			w.logger.Error(err, "publishing write to after writes topic")
		}
	}

	return nil
}
