package workers

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// PreUpdatesWorker updates data from the pending updates topic to the database.
type PreUpdatesWorker struct {
	logger               logging.Logger
	tracer               tracing.Tracer
	encoder              encoding.ClientEncoder
	postUpdatesPublisher publishers.Publisher
	dataManager          database.DataManager
	itemsIndexManager    search.IndexManager
}

// ProvidePreUpdatesWorker provides a PreUpdatesWorker.
func ProvidePreUpdatesWorker(
	ctx context.Context,
	logger logging.Logger,
	client *http.Client,
	dataManager database.DataManager,
	postUpdatesPublisher publishers.Publisher,
	searchIndexLocation search.IndexPath,
	searchIndexProvider search.IndexManagerProvider,
) (*PreUpdatesWorker, error) {
	const name = "pre_updates"

	itemsIndexManager, err := searchIndexProvider(ctx, logger, client, searchIndexLocation, "items", "name", "details")
	if err != nil {
		return nil, fmt.Errorf("setting up items search index manager: %w", err)
	}

	w := &PreUpdatesWorker{
		logger:               logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:               tracing.NewTracer(name),
		encoder:              encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postUpdatesPublisher: postUpdatesPublisher,
		dataManager:          dataManager,
		itemsIndexManager:    itemsIndexManager,
	}

	return w, nil
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
	logger := w.logger.WithValue("data_type", msg.DataType)

	logger.Debug("message read")

	switch msg.DataType {
	case types.ItemDataType:
		if err := w.dataManager.UpdateItem(ctx, msg.Item); err != nil {
			return observability.PrepareError(err, logger, span, "creating item")
		}

		if err := w.itemsIndexManager.Index(ctx, msg.Item.ID, msg.Item); err != nil {
			return observability.PrepareError(err, logger, span, "indexing the item")
		}

		if w.postUpdatesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				Item:                    msg.Item,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}

			if err := w.postUpdatesPublisher.Publish(ctx, dcm); err != nil {
				return observability.PrepareError(err, logger, span, "publishing data change message")
			}
		}
	case types.UserMembershipDataType, types.WebhookDataType:
		break
	}

	return nil
}
