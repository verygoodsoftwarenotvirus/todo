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

// ErrorMessage represents an event that asks a worker to write data to the datastore.
type ErrorMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// ErrorsWorker writes data from the pending writes topic to the database.
type ErrorsWorker struct {
	logger         logging.Logger
	tracer         tracing.Tracer
	errorsProducer events.Producer
	dataManager    database.DataManager
}

// ProvideErrorsWorker provides a ErrorsWorker.
func ProvideErrorsWorker(logger logging.Logger, dataManager database.DataManager, errorsProducer events.Producer) *ErrorsWorker {
	name := "errors_worker"
	return &ErrorsWorker{
		logger:         logging.EnsureLogger(logger).WithName(name),
		tracer:         tracing.NewTracer(name),
		errorsProducer: errorsProducer,
		dataManager:    dataManager,
	}
}

// HandleMessage handles a pending write.
func (w *ErrorsWorker) HandleMessage(message *nsq.Message) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *ErrorMessage

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
	}

	return nil
}
