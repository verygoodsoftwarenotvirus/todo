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
func (w *PreWritesWorker) HandleMessage(ctx context.Context, message []byte) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	var msg *types.PreWriteMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributableToUserID)

	w.logger.WithValue("data_type", msg.DataType).Debug("message read")

	switch msg.DataType {
	case types.ItemDataType:
		item, err := w.dataManager.CreateItem(ctx, msg.Item)
		if err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		} else if w.postWritesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				Item:                    item,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}
			if err = w.postWritesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-writes topic")
			}
		}
	case types.WebhookDataType:
		webhook, err := w.dataManager.CreateWebhook(ctx, msg.Webhook)
		if err != nil {
			return observability.PrepareError(err, w.logger, span, "creating webhook")
		} else if w.postWritesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				Webhook:                 webhook,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}
			if err = w.postWritesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-writes topic")
			}
		}
	case types.UserMembershipDataType:
		if err := w.dataManager.AddUserToAccount(ctx, msg.UserMembership); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating webhook")
		} else if w.postWritesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}
			if err = w.postWritesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-writes topic")
			}
		}
	}

	return nil
}
