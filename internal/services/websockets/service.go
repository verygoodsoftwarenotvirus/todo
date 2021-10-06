package websockets

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/consumers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	serviceName          = "websockets_service"
	dataChangesTopicName = "data_changes"
)

type (
	websocketConnection interface {
		SetWriteDeadline(t time.Time) error
		WriteMessage(messageType int, data []byte) error
		WriteControl(messageType int, data []byte, deadline time.Time) error
	}

	// service handles websockets.
	service struct {
		logger                      logging.Logger
		encoderDecoder              encoding.ServerEncoderDecoder
		tracer                      tracing.Tracer
		connections                 map[string][]websocketConnection
		sessionContextDataFetcher   func(*http.Request) (*types.SessionContextData, error)
		websocketConnectionUpgrader websocket.Upgrader
		cookieName                  string
		websocketDeadline           time.Duration
		pollDuration                time.Duration
		connectionsHat              sync.RWMutex
	}
)

// ProvideService builds a new websocket service.
func ProvideService(
	ctx context.Context,
	authCfg *authservice.Config,
	logger logging.Logger,
	encoder encoding.ServerEncoderDecoder,
	consumerProvider consumers.ConsumerProvider,
) (types.WebsocketDataService, error) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		Error:            buildWebsocketErrorFunc(encoder),
	}

	svc := &service{
		logger:                      logging.EnsureLogger(logger).WithName(serviceName),
		sessionContextDataFetcher:   authservice.FetchContextFromRequest,
		encoderDecoder:              encoder,
		websocketConnectionUpgrader: upgrader,
		cookieName:                  authCfg.Cookies.Name,
		connections:                 map[string][]websocketConnection{},
		websocketDeadline:           5 * time.Second,
		pollDuration:                10 * time.Second,
		tracer:                      tracing.NewTracer(serviceName),
	}

	dataChangesConsumer, err := consumerProvider.ProviderConsumer(ctx, dataChangesTopicName, svc.handleDataChange)
	if err != nil {
		return nil, fmt.Errorf("setting up event publisher: %w", err)
	}

	go svc.pingConnections()
	go dataChangesConsumer.Consume(nil, nil)

	return svc, nil
}

func buildWebsocketErrorFunc(encoder encoding.ServerEncoderDecoder) func(http.ResponseWriter, *http.Request, int, error) {
	return func(res http.ResponseWriter, req *http.Request, status int, reason error) {
		encoder.EncodeErrorResponse(req.Context(), res, reason.Error(), status)
	}
}
