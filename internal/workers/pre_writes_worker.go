package workers

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"

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
	itemsIndexManager   search.IndexManager
}

// ProvidePreWritesWorker provides a PreWritesWorker.
func ProvidePreWritesWorker(
	ctx context.Context,
	logger logging.Logger,
	dataManager database.DataManager,
	postWritesPublisher publishers.Publisher,
	searchIndexLocation search.IndexPath,
	searchIndexProvider search.IndexManagerProvider,
) (*PreWritesWorker, error) {
	const name = "pre_writes"

	itemsIndexManager, err := searchIndexProvider(ctx, logger, searchIndexLocation, "items", "name", "description")
	if err != nil {
		return nil, fmt.Errorf("setting up items search index manager: %w", err)
	}

	w := &PreWritesWorker{
		logger:              logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:              tracing.NewTracer(name),
		encoder:             encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postWritesPublisher: postWritesPublisher,
		dataManager:         dataManager,
		itemsIndexManager:   itemsIndexManager,
	}

	return w, err
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
				observability.AcknowledgeError(err, w.logger, span, "publishing to post-writes topic")
			}

			if err = w.itemsIndexManager.Index(ctx, item.ID, item); err != nil {
				observability.AcknowledgeError(err, w.logger, span, "indexing the item")
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
