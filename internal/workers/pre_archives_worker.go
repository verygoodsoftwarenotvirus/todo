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

// PreArchivesWorker archives data from the pending archives topic to the database.
type PreArchivesWorker struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	encoder               encoding.ClientEncoder
	postArchivesPublisher publishers.Publisher
	dataManager           database.DataManager
	itemsIndexManager     search.IndexManager
}

// ProvidePreArchivesWorker provides a PreArchivesWorker.
func ProvidePreArchivesWorker(
	ctx context.Context,
	logger logging.Logger,
	client *http.Client,
	dataManager database.DataManager,
	postArchivesPublisher publishers.Publisher,
	searchIndexLocation search.IndexPath,
	searchIndexProvider search.IndexManagerProvider,
) (*PreArchivesWorker, error) {
	const name = "pre_archives"

	itemsIndexManager, err := searchIndexProvider(ctx, logger, client, searchIndexLocation, "items", "name", "description")
	if err != nil {
		return nil, fmt.Errorf("setting up items search index manager: %w", err)
	}

	w := &PreArchivesWorker{
		logger:                logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:                tracing.NewTracer(name),
		encoder:               encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postArchivesPublisher: postArchivesPublisher,
		dataManager:           dataManager,
		itemsIndexManager:     itemsIndexManager,
	}

	return w, nil
}

// HandleMessage handles a pending archive.
func (w *PreArchivesWorker) HandleMessage(ctx context.Context, message []byte) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	var msg *types.PreArchiveMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributableToUserID)
	logger := w.logger.WithValue("data_type", msg.DataType)

	logger.Debug("message read")

	switch msg.DataType {
	case types.ItemDataType:
		if err := w.dataManager.ArchiveItem(ctx, msg.ItemID, msg.AttributableToAccountID); err != nil {
			return observability.PrepareError(err, w.logger, span, "archiving item")
		}

		if err := w.itemsIndexManager.Delete(ctx, msg.ItemID); err != nil {
			return observability.PrepareError(err, w.logger, span, "removing item from index")
		}

		if w.postArchivesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}

			if err := w.postArchivesPublisher.Publish(ctx, dcm); err != nil {
				return observability.PrepareError(err, logger, span, "publishing data change message")
			}
		}
	case types.WebhookDataType:
		if err := w.dataManager.ArchiveWebhook(ctx, msg.WebhookID, msg.AttributableToAccountID); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postArchivesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}

			if err := w.postArchivesPublisher.Publish(ctx, dcm); err != nil {
				return observability.PrepareError(err, logger, span, "publishing data change message")
			}
		}
	case types.UserMembershipDataType:
		break
	}

	return nil
}
